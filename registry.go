package registry

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

//Registry Register, Unregister
type Registry interface {
	Register(s Service) (FnUnregister, error)
	GetServices(name string) ([]Service, error)
	GetService(name string, filter Filter) (*Service, error)
	Observe(serviceName string) error
	Close() error
}

//Pong response to ping
type Pong struct {
	Service
}

//FnUnregister call this func for unregister the service
type FnUnregister func()

//Service service struct
type Service struct {
	//Network tcp/unix/tpc6
	Network string
	//Bind address
	Address string
	//URL base used for communicate with this service
	URL string
	//Name service name
	Name string
	//Version semver version
	Version string
	//Host
	Host string
}

type observe struct {
	callback func(pong Pong)
}

type reg struct {
	messageBase string
	m           map[string]map[string]Pong
	observers   map[string]observe
	opts        Options

	//manage Registerd service
	registeredServices map[string]Pong
	//used to stop channel
	chStopChannelRegisteredServices chan bool
	chFiredRegisteredService        chan Pong
}

var (
	//ErrNotFound when no service found
	ErrNotFound = errors.New("No service found")
	//singleton instance
	instance    *reg
	intanceOnce sync.Once
)

func (r reg) buildMessage(message, service string, args ...string) string {
	var b strings.Builder
	b.WriteString(r.messageBase)
	b.WriteString("/")
	b.WriteString(message)
	b.WriteString("/")
	b.WriteString(service)
	if args != nil {
		for _, s := range args {
			b.WriteString("/")
			b.WriteString(s)
		}
	}
	return r.messageBase + "/" + message + "/" + service
}

func (r reg) subToPing(s Service) {
	log.Info("Sub to ping for service ", s.Name, " ", s.Address)
	pong := Pong{
		Service: s,
	}
	fn := func(m *nats.Msg) {
		data, err := json.Marshal(pong)
		if err == nil {
			log.Info("Respond to ping ", m.Reply, " ", s.Name, " ", s.Address)
			m.Respond(data)
		} else {
			log.Errorf("Unable to marchal pong for service %s error is: %s", s.Name, err)
		}

	}
	r.opts.natsConn.Subscribe(r.buildMessage("ping", s.Name), fn)
}

func (r reg) Register(s Service) (f FnUnregister, err error) {
	f = func() {}
	pong := Pong{Service: s}

	r.subToPing(s)

	r.registeredServices[s.Name+s.Address] = pong
	//notify the channel to send new message
	r.chFiredRegisteredService <- pong
	f = func() {
		r.Unregister(s)
	}
	return

}

func (r reg) pubregister(pong Pong) (err error) {
	var data []byte
	if data, err = json.Marshal(pong); err != nil {
		log.Error("publish register failed unmarshal service ", pong.Name, " :", err)
		return
	}
	if err = r.opts.natsConn.Publish(r.buildMessage("register", pong.Name), data); err != nil {
		log.Error("publish register failed for service ", pong.Name, " :", err)
		return
	}
	log.Info("Send register for service ", pong.Host, " ", pong.Service)
	return
}

func (r reg) registerServiceInContinue() {
	log.Infof("Start go routine for register services every %s ", r.opts.registerInterval)
stop:
	for {
		tk := time.Tick(r.opts.registerInterval)
		select {
		case <-r.chStopChannelRegisteredServices:
			log.Info("Receive stop in channel")
			break stop
		case <-tk:
			for _, pong := range r.registeredServices {
				r.pubregister(pong)
			}
		case pong := <-r.chFiredRegisteredService:
			r.pubregister(pong)
		}
	}
	log.Info("Stop go routine registerSerivceInContinue")
}

func (r reg) Unregister(s Service) (err error) {
	var data []byte
	if data, err = json.Marshal(s); err != nil {
		return
	}
	err = r.opts.natsConn.Publish(r.buildMessage("unregister", s.Name), data)
	if r.registeredServices != nil {
		delete(r.registeredServices, s.Name+s.Address)
	}
	log.Info("service ", s.Host, " ", s.Name, " Unregistered")
	return

}

//Connect pubsub transport
func Connect(opts ...Option) (r Registry, err error) {
	if instance == nil {
		intanceOnce.Do(func() {
			instance = &reg{
				messageBase:                     "registry",
				m:                               make(map[string]map[string]Pong),
				observers:                       make(map[string]observe),
				opts:                            newOptions(opts...),
				registeredServices:              make(map[string]Pong),
				chFiredRegisteredService:        make(chan Pong),
				chStopChannelRegisteredServices: make(chan bool),
			}
			go instance.registerServiceInContinue()
		})
	}
	r = instance
	return
}

func (r reg) getServices(name string) (res []Service) {
	var servicesMap map[string]Pong
	var ok bool
	if servicesMap, ok = r.m[name]; ok && len(servicesMap) > 0 {
		for _, v := range servicesMap {
			res = append(res, v.Service)
		}
		return
	}
	return
}

func (r reg) GetService(name string, f Filter) (*Service, error) {
	services, err := r.getinternalService(name, f)
	if err != nil {
		return nil, err
	}
	if services == nil || len(services) == 0 {
		return nil, ErrNotFound
	}
	//return first item
	return &services[0], nil

}

func (r reg) getinternalService(name string, f Filter) ([]Service, error) {
	res := r.getServices(name)
	if res != nil {
		if f != nil {
			for _, s := range res {
				if f(s) {
					return []Service{s}, nil
				}
			}
			return nil, ErrNotFound
		}
		return res, nil
	}
	var observe *observe
	if o, exists := r.observers[name]; !exists {
		observe = &o
		r.Observe(name)
	}
	err := r.opts.natsConn.PublishRequest(r.buildMessage("ping", name), r.buildMessage("register", name), nil)
	if err != nil {
		return nil, err
	}
	ch := make(chan Pong)
	if f != nil {
		observe.callback = func(pong Pong) {
			if f(pong.Service) {
				ch <- pong
			}
		}
	}

	//Waiting for context done
	tk := time.Tick(r.opts.timeout)
	select {
	case <-tk:
		break
	case <-ch:
		close(ch)
		break
	}

	res = r.getServices(name)

	return res, nil

}

func (r reg) GetServices(name string) ([]Service, error) {
	return r.getinternalService(name, nil)
}

func (r reg) append(s Pong) {
	if _, ok := r.m[s.Name]; !ok {
		r.m[s.Name] = make(map[string]Pong)
	}
	services := r.m[s.Name]
	services[s.Host+s.Address] = s
}

func (r reg) subregister(msg *nats.Msg) {
	var s Pong
	err := json.Unmarshal(msg.Data, &s)
	if err != nil {
		logrus.Errorf("unmarshal error when sub to register: %s", err)
		return
	}
	r.append(s)
}
func (r reg) subunregister(msg *nats.Msg) {
	var s Service
	err := json.Unmarshal(msg.Data, &s)
	if err != nil {
		logrus.Errorf("unmarshal error when sub to register: %s", err)
		return
	}

	if _, ok := r.m[s.Name]; !ok {
		return
	}
	delete(r.m[s.Name], s.Host+s.Address)

}
func (r reg) Observe(service string) error {
	_, err := r.opts.natsConn.Subscribe(r.buildMessage("register", service), r.subregister)
	if err != nil {
		return err
	}
	_, err = r.opts.natsConn.Subscribe(r.buildMessage("unregister", service), r.subunregister)
	if err != nil {
		return err
	}
	r.observers[service] = observe{}
	return nil
}

//Close unregister to all subscriptions.
//Clear local cache.
//Stop go routine if exist.
//TODO
func (r reg) Close() (err error) {
	r.chStopChannelRegisteredServices <- true
	for _, s := range r.registeredServices {
		r.Unregister(s.Service)
	}
	log.Info("Close registry")
	return
}
