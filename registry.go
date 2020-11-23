package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

//Registry Register, Unregister
type Registry interface {
	Register(s Service) (FnUnregister, error)
	Unregister(s Service) error
	GetServices(name string) ([]Service, error)
	GetService(name string, filters ...Filter) (*Service, error)
	Observe(serviceName string) error
	Close() error
}

//Timestamps define registered datetime and expiration duration
type Timestamps struct {
	//Registered date registered in nanoseconds (unix timestamp in milliseconds)
	Registered int64
	//expiration duration in milliseconds
	Duration int
}

//Pong response to ping
type Pong struct {
	Service
	Timestamps *Timestamps `json:"t,omitempty"`
}

//Event represent event (register|unregister|unavailbale)
type Event string

const (
	//Register register event
	Register Event = "register"
	//Unregister unregister event
	Unregister Event = "unregister"
)

//FnUnregister call this func for unregister the service
type FnUnregister func()

//ObserverEvent event tigered
type ObserverEvent func(s Service, ev Event)

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
	callback func(pong *Pong)
}

type reg struct {
	ser       *services
	observers map[string]*observe
	opts      Options

	//manage Registered service
	registeredServices map[string]*Pong

	//used to stop go routine registerServiceInContinue
	chStopChannelRegisteredServices chan bool

	//wake up go routine registerServiceInContinue and force a new service
	//to be registered
	chFiredRegisteredService chan *Pong

	//references to all subscriptions to nats
	//these subscriptions unsubscripe when Close function will be call
	subscriptions []Subscription
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
	b.WriteString(".")
	b.WriteString(message)
	b.WriteString(".")
	b.WriteString(service)
	if args != nil {
		for _, s := range args {
			b.WriteString(".")
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

func (r reg) subToPing(p *Pong) {
	log.Info("Sub to ping for service ", p.Name, " ", p.Address)
	fn := func(m *PubsubMsg) {
		r.pubregister(p)
	}
	s, err := r.opts.pubsub.Sub(r.buildMessage("ping", p.Name), fn)
	if err != nil {
		log.Error("subToPing: ", err)
	} else {
		log.Debugf("Subscribe for %s OK", s.Subject())
		r.subscriptions = append(r.subscriptions, s)
	}
}

func (r reg) Register(s Service) (f FnUnregister, err error) {
	if s.Host == "" {
		s.Host = r.opts.hostname
	}
	p := &Pong{Service: s, Timestamps: &Timestamps{Registered: time.Now().UnixNano(), Duration: int(r.opts.registerInterval.Milliseconds())}}
	r.subToPing(p)

	r.registeredServices[s.Name+s.Address] = p
	//notify the channel to send new message
	r.chFiredRegisteredService <- p
	f = func() {
		r.Unregister(s)
	}
	return

}

func (r reg) pubregister(p *Pong) (err error) {
	var data []byte
	p.Timestamps.Registered = time.Now().UnixNano() / 1000000
	if data, err = json.Marshal(p); err != nil {
		log.Error("publish register failed unmarshal service ", p.Name, " :", err)
		return
	}
	topic := r.buildMessage("register", p.Name)
	if err = r.opts.pubsub.Pub(topic, data); err != nil {
		log.Error("publish register failed for service ", p.Name, " :", err)
		return
	}
	log.Infof("%s (%s) host: %s", topic, p.Address, p.Host)
	return
}

func (r reg) registerServiceInContinue() {
	log.Infof("Start go routine for register services every %s ", r.opts.registerInterval)
	tk := time.NewTicker(r.opts.registerInterval)
	tkDue := time.NewTicker(r.opts.checkDueTime)
stop:
	for {
		select {
		case <-r.chStopChannelRegisteredServices:
			log.Info("Receive stop in channel")
			break stop
		case <-tk.C:
			for _, pong := range r.registeredServices {
				r.pubregister(pong)
			}
		case pong := <-r.chFiredRegisteredService:
			r.pubregister(pong)
		case <-tkDue.C:
			r.checkDueTime()
		}
	}
	tk.Stop()
	tkDue.Stop()
	log.Info("Stop go routine registerSerivceInContinue")
}

func (r reg) checkDueTime() {
	toDel := []*Pong{}
	now := time.Now()
	r.ser.IterateAll(func(key string, s *Pong) bool {
		if now.After(s.dueTime) {
			toDel = append(toDel, s)
		}
		return true
	})

	toRefresh := make(map[string]bool)
	for _, k := range toDel {
		r.ser.Delete(k)
		toRefresh[k.Name] = true
	}
	for k := range toRefresh {
		r.ser.rebuildCache(k)
	}
}

func (r reg) Unregister(s Service) (err error) {
	var data []byte
	if data, err = json.Marshal(s); err != nil {
		return
	}
	topic := r.buildMessage("unregister", s.Name)
	err = r.opts.pubsub.Pub(topic, data)
	if r.registeredServices != nil {
		delete(r.registeredServices, s.Name+s.Address)
	}
	log.Infof("%s (%s) host: %s", topic, s.Address, s.Host)
	return

}

//Connect pubsub transport
func Connect(opts ...Option) (r Registry, err error) {
	if instance == nil {
		intanceOnce.Do(func() {
			instance = &reg{
				ser:                             newServices(),
				observers:                       make(map[string]*observe),
				opts:                            newOptions(opts...),
				registeredServices:              make(map[string]*Pong),
				chFiredRegisteredService:        make(chan *Pong),
				chStopChannelRegisteredServices: make(chan bool),
				subscriptions:                   make([]Subscription, 0),
			}
			go instance.registerServiceInContinue()
		})
	}
	r = instance
	return
}

func (r reg) GetService(name string, f ...Filter) (*Service, error) {
	services, err := r.getinternalService(name, f...)
	if err != nil {
		return nil, err
	}
	if services == nil || len(services) == 0 {
		return nil, ErrNotFound
	}
	//return first item
	return &services[0], nil

}

func chainFilters(pongs []*Pong, filters ...Filter) []Service {
	services := []*Pong{}
	for _, v := range pongs {
		services = append(services, v)
	}
	if filters != nil {
		for _, f := range filters {
			if f != nil {
				services = f(services)
			}
		}
	}

	res := make([]Service, len(services))
	for i, p := range services {
		res[i] = p.Service
	}
	return res
}

func (r reg) getinternalService(name string, serviceFilters ...Filter) ([]Service, error) {
	filters := serviceFilters
	if filters == nil {
		filters = r.opts.filters
	}
	//service is already registered

	if res := r.ser.GetServices(name); res != nil {
		return chainFilters(res, filters...), nil
	}
	//service not yet registered
	//register invoke r.Observe(service) with a callback containing a channel
	//the callback apply filters on service and write in the channel when a service is ok with the filters
	ch := make(chan *Service)
	observe := &observe{}
	r.observers[name] = observe
	if filters != nil {
		observe.callback = func(p *Pong) {
			ok := true
			arg := []*Pong{p}
			for _, f := range filters {
				if filtered := f(arg); filtered == nil || len(filtered) == 0 {
					ok = false
				}
			}
			if ok {
				observe.callback = nil
				ch <- &p.Service
			}
		}
	} else {
		observe.callback = func(p *Pong) {
			observe.callback = nil
			ch <- &p.Service
		}
	}
	r.Observe(name)
	err := r.opts.pubsub.Pub(r.buildMessage("ping", name), nil)
	if err != nil {
		return nil, err
	}

	//create timeout if no service available
	var serviceFound *Service
	tk := time.NewTimer(r.opts.timeout)
	select {
	case <-tk.C:
		break
	case serviceFound = <-ch:
		close(ch)
		break
	}
	if serviceFound != nil {
		return []Service{*serviceFound}, nil
	}
	tk.Stop()
	return nil, ErrNotFound

}

func (r reg) GetServices(name string) ([]Service, error) {
	return r.getinternalService(name)
}

func (r reg) subregister(msg *PubsubMsg) {
	var p *Pong
	err := json.Unmarshal(msg.Data, &p)
	if err != nil {
		logrus.Errorf("unmarshal error when sub to register: %s", err)
		return
	}
	for _, f := range r.opts.observeFilters {
		if !f(p) {
			return
		}
	}
	if o, ok := r.observers[p.Name]; ok {
		if o.callback != nil {
			o.callback(p)
		}
	} else {
		r.observers[p.Name] = &observe{}
	}
	var alreadyExist bool
	if p, alreadyExist = r.ser.LoadOrStore(p); !alreadyExist && r.opts.observerEvent != nil {
		r.opts.observerEvent(p.Service, Register)
	}
	if p.Timestamps != nil {
		d := int(float32(p.Timestamps.Duration) * r.opts.dueDurationFactor)
		registered := p.Timestamps.Registered * int64(time.Millisecond)
		p.dueTime = time.Unix(0, registered).Add(time.Duration(d) * time.Millisecond)
		log.Debug(p.dueTime.Local().Format(time.ANSIC))
	}
	log.Debugf("append %s ", p.Service)

}

func (r reg) subunregister(msg *PubsubMsg) {
	var s Service
	err := json.Unmarshal(msg.Data, &s)
	if err != nil {
		logrus.Errorf("unmarshal error when sub to register: %s", err)
		return
	}
	p := &Pong{Service: s}
	for _, f := range r.opts.observeFilters {
		if !f(p) {
			return
		}
	}
	if r.opts.observerEvent != nil {
		r.opts.observerEvent(s, Unregister)
	}
	r.ser.DeleteByName(s.Name + s.Address)
	logrus.Debugf("Unregister service %s/%s", s.Name, s.Address)
}

func (r reg) Observe(service string) error {
	if _, ok := r.observers[service]; !ok {
		r.observers[service] = &observe{}
	}
	s, err := r.opts.pubsub.Sub(r.buildMessage("register", service), r.subregister)
	if err != nil {
		return err
	}
	r.subscriptions = append(r.subscriptions, s)
	s, err = r.opts.pubsub.Sub(r.buildMessage("unregister", service), r.subunregister)
	if err != nil {
		return err
	}
	r.subscriptions = append(r.subscriptions, s)
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
		s.Unsub()
	}
	r.subscriptions = r.subscriptions[0:0]
	instance = nil
	intanceOnce = sync.Once{}
	r.opts.pubsub.Stop()
	log.Debug("Close registry done")
	return
}
