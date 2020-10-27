package registry

import (
	"encoding/json"
	"errors"
	"fmt"
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

//Timestamps define registered datetime and expiration duration
type Timestamps struct {
	//date registered in nanoseconds (unix timestamp)
	Registered int64
	//expiration duration in milliseconds
	Duration int
}

//Pong response to ping
type Pong struct {
	Service
	Timestamps *Timestamps `json:"t,omitempty"`
}

//FnUnregister call this func for unregister the service
type FnUnregister func()

//Service service struct
type Service struct {
	//Network tcp/unix/tpc6
	Network string `json:"net,omitempty"`
	//Bind address
	Address string `json:"add,omitempty"`
	//URL base used for communicate with this service
	URL string `json:"url,omitempty"`
	//Name service name
	Name string `json:"name,omitempty"`
	//Version semver version
	Version string `json:"v,omitempty"`
	//Host
	Host string `json:"h,omitempty"`

	dueTime time.Time
}

//DueTime expiration time
func (s Service) DueTime() time.Time {
	return s.dueTime
}

type observe struct {
	callback func(pong Pong)
}

type reg struct {
	m         map[string]map[string]Pong
	observers map[string]observe
	opts      Options

	//manage Registerd service
	registeredServices map[string]Pong

	//used to stop go routine registerServiceInContinue
	chStopChannelRegisteredServices chan bool

	//wakeup go routine registerServiceInContinue and force a new service
	//to be registered
	chFiredRegisteredService chan Pong

	//references to all subscriptions to nats
	//these subscriptions unsubscripe when Close function will be call
	subscriptions []*nats.Subscription
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
	b.WriteString(r.opts.mainTopic)
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
	return b.String()
}

func (s Service) String() string {
	due := s.DueTime().Sub(time.Now())
	return fmt.Sprintf("Name: %s Addr: %s Host: %s URL: %s Timestamps in due time in %d millis", s.Name, s.Address, s.Host, s.URL, int(due.Milliseconds()))
}

func (p Pong) String() string {
	if p.Timestamps == nil {
		return fmt.Sprintf("%s", p.Name)
	}
	return fmt.Sprintf("%s timestamp %d, %d", p.Name, p.Timestamps.Registered, p.Timestamps.Duration)
}

func (r reg) subToPing(p Pong) {
	log.Info("Sub to ping for service ", p.Name, " ", p.Address)
	fn := func(m *nats.Msg) {
		data, err := json.Marshal(p)
		if err == nil {
			log.Info("Respond to ping ", m.Reply, " ", p.Name, " ", p.Address)
			m.Respond(data)
		} else {
			log.Errorf("Unable to marchal pong for service %s error is: %s", p.Name, err)
		}

	}
	s, err := r.opts.natsConn.Subscribe(r.buildMessage("ping", p.Name), fn)
	if err != nil {
		log.Error("subToPing: ", err)
	} else {
		r.subscriptions = append(r.subscriptions, s)
	}
}

func (r reg) Register(s Service) (f FnUnregister, err error) {
	f = func() {}
	p := Pong{Service: s, Timestamps: &Timestamps{Registered: time.Now().UnixNano(), Duration: int(r.opts.registerInterval.Milliseconds())}}
	r.subToPing(p)

	r.registeredServices[s.Name+s.Address] = p
	//notify the channel to send new message
	r.chFiredRegisteredService <- p
	f = func() {
		r.Unregister(s)
	}
	return

}

func (r reg) pubregister(p Pong) (err error) {
	var data []byte
	p.Timestamps.Registered = time.Now().UnixNano()
	if data, err = json.Marshal(p); err != nil {
		log.Error("publish register failed unmarshal service ", p.Name, " :", err)
		return
	}
	if err = r.opts.natsConn.Publish(r.buildMessage("register", p.Name), data); err != nil {
		log.Error("publish register failed for service ", p.Name, " :", err)
		return
	}
	log.Info("Send register for service ", p.Host, " ", p)
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
				m:                               make(map[string]map[string]Pong),
				observers:                       make(map[string]observe),
				opts:                            newOptions(opts...),
				registeredServices:              make(map[string]Pong),
				chFiredRegisteredService:        make(chan Pong),
				chStopChannelRegisteredServices: make(chan bool),
				subscriptions:                   make([]*nats.Subscription, 0),
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
			s := v.Service
			res = append(res, s)
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
		observe.callback = func(p Pong) {
			if f(p.Service) {
				ch <- p
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

func (r reg) append(p Pong) {
	if _, ok := r.m[p.Name]; !ok {
		r.m[p.Name] = make(map[string]Pong)
	}
	services := r.m[p.Name]
	log.Debugf("append %s", p)
	if p.Timestamps != nil {
		p.dueTime = time.Unix(0, p.Timestamps.Registered).Add(time.Duration(p.Timestamps.Duration) * time.Millisecond)
	}

	services[p.Host+p.Address] = p
}

func (r reg) subregister(msg *nats.Msg) {
	var s Pong
	deb := string(msg.Data)
	if len(deb) == 0 {

	}
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
	s, err := r.opts.natsConn.Subscribe(r.buildMessage("register", service), r.subregister)
	if err != nil {
		return err
	}
	r.subscriptions = append(r.subscriptions, s)
	s, err = r.opts.natsConn.Subscribe(r.buildMessage("unregister", service), r.subunregister)
	if err != nil {
		return err
	}
	r.subscriptions = append(r.subscriptions, s)
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
	for _, s := range r.subscriptions {
		s.Unsubscribe()
	}
	r.subscriptions = r.subscriptions[0:0]
	log.Debug("Close registry done")
	return
}
