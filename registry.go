package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/eddieraa/registry/pubsub"
	"github.com/sirupsen/logrus"
)

// Registry Register, Unregister
// Registry defines an interface for a service registry that allows for
// registering, unregistering, and querying services. It also provides
// methods for observing services and managing their statuses.
//
// Methods:
//   - Register(s Service) (FnUnregister, error): Registers a service and returns
//     a function to unregister the service along with any error encountered.
//   - Unregister(s Service) error: Unregisters a service and returns any error
//     encountered.
//   - GetServices(name string, options ...func(*getServicesOptions)) ([]Service, error):
//     Retrieves a list of services by name with optional filtering options.
//   - GetService(name string, filters ...Filter) (*Service, error): Retrieves a
//     single service by name with optional filters.
//   - Observe(serviceName string) error: Starts observing a service by name and
//     returns any error encountered.
//   - GetObservedServiceNames() []string: Returns a list of names of observed
//     services.
//   - Subscribers() []string: Returns a list of subscribers.
//   - Close() error: Closes the registry and returns any error encountered.
//   - SetServiceStatus(s Service, status Status) error: Sets the status of a
//     service and returns any error encountered.
//   - GetRegisteredServices() []Service: Returns a list of all registered services.
type Registry interface {
	Register(s Service) (FnUnregister, error)
	Unregister(s Service) error
	GetServices(name string, options ...func(*getServicesOptions)) ([]Service, error)
	GetService(name string, filters ...Filter) (*Service, error)
	Observe(serviceName string) error
	GetObservedServiceNames() []string
	Subscribers() []string
	Close() error
	SetServiceStatus(s Service, status Status) error
	GetRegisteredServices() []Service
}

type Status int

const (
	Passing Status = iota
	Warning
	Critical
)

func (s Status) String() string {
	return [...]string{"passing", "warning", "critical"}[s]
}
func (s Status) FromString(status string) Status {
	return map[string]Status{"": Passing, "passing": Passing, "warning": Warning, "critical": Critical}[status]
}
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
func (s *Status) UnmarshalJSON(b []byte) error {
	var statusString string
	if err := json.Unmarshal(b, &statusString); err != nil {
		return err
	}
	*s = s.FromString(statusString)
	return nil
}

// Timestamps define registered datetime and expiration duration
type Timestamps struct {
	//Registered date registered in nanoseconds (unix timestamp in milliseconds)
	Registered int64
	//expiration duration in milliseconds
	Duration int
}

// Pong response to ping
type Pong struct {
	Service
	Timestamps *Timestamps `json:"t,omitempty"`
	Status     Status      `json:"status,omitempty"`
}

// Event represent event (register|unregister|unavailbale)
type Event string

const (
	//EventRegister register event
	EventRegister Event = "register"
	//EventUnregister unregister event
	EventUnregister Event = "unregister"
)

// FnUnregister call this func for unregister the service
type FnUnregister func()

// ObserverEvent event tigered
type ObserverEvent func(s Service, ev Event)

// Service service struct
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
	//KV key value pair
	KV map[string]string `json:"kv,omitempty"`

	dueTime time.Time
}

// DueTime expiration time
func (s Service) DueTime() time.Time {
	return s.dueTime
}

type observe struct {
	callback func(pong *Pong)
}

type reg struct {
	ser         *services
	_observers  map[string]*observe
	observersMu sync.Mutex
	opts        Options

	//manage Registered service
	//registeredServices    map[string]*Pong
	registeredServicesMap sync.Map

	//used to stop go routine registerServiceInContinue
	chStopChannelRegisteredServices chan bool

	//wake up go routine registerServiceInContinue and force a new service
	//to be registered
	chFiredRegisteredService chan *Pong

	//references to all subscriptions to nats
	//these subscriptions unsubscripe when Close function will be call
	subscriptions []pubsub.Subscription

	log *logrus.Logger
}

type getServicesOptions struct {
	nowait  bool
	filters []Filter
	ctx     context.Context
}

func NoWait() func(*getServicesOptions) {
	return func(gso *getServicesOptions) {
		gso.nowait = true
	}
}

func WithFilters(filters ...Filter) func(*getServicesOptions) {
	return func(gso *getServicesOptions) {
		gso.filters = filters
	}
}
func WithContext(ctx context.Context) func(*getServicesOptions) {
	return func(gso *getServicesOptions) {
		gso.ctx = ctx
	}
}

var (
	//ErrNotFound when no service found
	ErrNotFound = errors.New("no service found")
	//ErrNoDefaultInstance when intance singleton has not been set
	ErrNoDefaultInstance = errors.New("default instance has not been set, call SetDefault before")
	//singleton instance
	instance *reg
	mu       sync.Mutex
)

func (r *reg) buildMessage(message, service string) string {
	var b strings.Builder
	b.WriteString(r.opts.mainTopic)
	b.WriteString(".")
	b.WriteString(message)
	b.WriteString(".")
	b.WriteString(service)
	return b.String()
}

func (s Service) String() string {
	due := time.Until(s.DueTime())
	return fmt.Sprintf("Name: %s Addr: %s Host: %s URL: %s Timestamps in due time in %d millis", s.Name, s.Address, s.Host, s.URL, int(due.Milliseconds()))
}

func (p Pong) String() string {
	if p.Timestamps == nil {
		return p.Name
	}
	return fmt.Sprintf("%s timestamp %d, %d", p.Name, p.Timestamps.Registered, p.Timestamps.Duration)
}

func (r *reg) subToPing(p *Pong) error {
	r.log.Info("Sub to ping for service ", p.Name, " ", p.Address)
	fn := func(m *pubsub.PubsubMsg) {
		r.pubregister(p)
	}
	s, err := r.opts.pubsub.Sub(r.buildMessage("ping", p.Name), fn)
	if err != nil {
		return err
	}
	r.log.Debugf("Subscribe for %s OK", s.Subject())
	r.subscriptions = append(r.subscriptions, s)
	return nil
}

func (r *reg) Register(s Service) (f FnUnregister, err error) {
	if s.Host == "" {
		s.Host = r.opts.hostname
	}
	p := &Pong{Service: s, Timestamps: &Timestamps{Registered: time.Now().UnixNano(), Duration: int(r.opts.registerInterval.Milliseconds())}}
	if err = r.subToPing(p); err != nil {
		r.log.Errorf("Failed to subscribe to ping for service %s: %v", s.Name, err)
		return nil, err
	}

	r.registeredServicesMap.Store(s.Name+s.Address, p)
	//notify the channel to send new message
	r.chFiredRegisteredService <- p
	f = func() {
		r.Unregister(s)
	}
	return

}

func (r *reg) Subscribers() []string {
	res := []string{}
	r.observersMu.Lock()
	for k := range r._observers {
		if !strings.HasSuffix(k, "*") {
			res = append(res, k)
		}
	}
	r.observersMu.Unlock()
	return res
}

func (r *reg) pubregister(p *Pong) (err error) {
	var data []byte
	p.Timestamps.Registered = time.Now().UnixNano() / 1000000
	data, err = json.Marshal(p)
	if err == nil {
		topic := r.buildMessage("register", p.Name)
		if err = r.opts.pubsub.Pub(topic, data); err != nil {
			r.log.Error("publish register failed for service ", p.Name, " :", err)
			return
		}
		r.log.Debugf("%s (%s) host: %s status %s", topic, p.Address, p.Host, p.Status)
	}
	return
}

func (r *reg) registerServiceInContinue() {
	r.log.Infof("Start go routine for register services every %s ", r.opts.registerInterval)
	tk := time.NewTicker(r.opts.registerInterval)
	tkDue := time.NewTicker(r.opts.checkDueTime)
stop:
	for {
		select {
		case <-r.chStopChannelRegisteredServices:
			r.log.Info("Receive stop in channel")
			break stop
		case <-tk.C:
			r.registeredServicesMap.Range(func(k, v interface{}) bool {
				r.pubregister(v.(*Pong))
				return true
			})
		case pong := <-r.chFiredRegisteredService:
			r.pubregister(pong)
		case <-tkDue.C:
			r.checkDueTime()
		}
	}
	tk.Stop()
	tkDue.Stop()
	r.log.Info("Stop go routine registerSerivceInContinue")
}

func (r *reg) checkDueTime() {
	toDel := []*Pong{}
	now := time.Now().Local()
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

func (r *reg) Unregister(s Service) (err error) {
	if _, ok := r.registeredServicesMap.Load(s.Name + s.Address); !ok {
		return ErrNotFound
	}
	var data []byte
	var topic string
	data, err = json.Marshal(s)
	if err == nil {
		topic = r.buildMessage("unregister", s.Name)
		err = r.opts.pubsub.Pub(topic, data)
		r.registeredServicesMap.Delete(s.Name + s.Address)
		r.log.Infof("%s (%s) host: %s", topic, s.Address, s.Host)
	}
	return
}

// GetObservedServiceNames return subscribed service names
func (r *reg) GetObservedServiceNames() (res []string) {
	r.observersMu.Lock()
	res = make([]string, len(r._observers))
	i := 0
	for k := range r._observers {
		if k != "*" {
			res[i] = k
			i++
		}
	}
	r.observersMu.Unlock()
	res = res[0:i]
	return
}
func (r *reg) SetServiceStatus(s Service, status Status) (err error) {
	if v, ok := r.registeredServicesMap.Load(s.Name + s.Address); ok {
		p := v.(*Pong)
		if p.Status != status {
			p.Status = status
			r.chFiredRegisteredService <- p
		}

	} else {
		err = ErrNotFound
	}
	return
}

func (r *reg) GetRegisteredServices() (services []Service) {
	r.registeredServicesMap.Range(func(k, v interface{}) bool {
		if services == nil {
			services = make([]Service, 0)
		}
		services = append(services, v.(*Pong).Service)
		return true
	})
	return
}

// NewRegistry create a new service registry instance
func NewRegistry(opts ...Option) (Registry, error) {
	o := newOptions(opts...)
	if cfg, ok := o.pubsub.(Configure); ok {
		if err := cfg.Configure(&o); err != nil {
			return nil, err
		}
	}
	log := logrus.New()
	r := &reg{
		ser:                             newServices(log),
		_observers:                      make(map[string]*observe),
		opts:                            o,
		chFiredRegisteredService:        make(chan *Pong),
		chStopChannelRegisteredServices: make(chan bool),
		subscriptions:                   make([]pubsub.Subscription, 0),
		log:                             log,
	}

	log.SetLevel(r.opts.loglevel)
	go r.registerServiceInContinue()
	return r, nil
}

// SetDefault set the default instance
//
// ex pubsub transport
func SetDefault(opts ...Option) (r Registry, err error) {
	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		r, err = NewRegistry(opts...)
		if err == nil {
			instance = r.(*reg)
		}
	}
	return
}

// GetDefault return default instance. return err if no default instance had been set
func GetDefault() (r Registry, err error) {
	if instance == nil {
		return nil, ErrNoDefaultInstance
	}
	return instance, nil
}

// GetService find service with service name
//
// Call SetDefault before use
func GetService(name string, f ...Filter) (*Service, error) {
	if instance == nil {
		return nil, ErrNoDefaultInstance
	}
	return instance.GetService(name, f...)
}

// Close the registry instance
//
// Call SetDefault before
func Close() error {
	if instance == nil {
		return ErrNoDefaultInstance
	}
	err := instance.Close()
	instance = nil
	return err
}

// GetServices return all registered service
//
// Call SetDefault before use
func GetServices(name string, opts ...func(*getServicesOptions)) ([]Service, error) {
	if instance == nil {
		return nil, ErrNoDefaultInstance
	}
	return instance.GetServices(name, opts...)
}

func GetObservedServiceNames() []string {
	if instance == nil {
		return nil
	}
	return instance.GetObservedServiceNames()
}

// Observe subscribe to service
//
// Call SetDefault before use
func Observe(name string) error {
	if instance == nil {
		return ErrNoDefaultInstance
	}
	return instance.Observe(name)
}

// Register register a new service
//
// Call SetDefault before use
func Register(s Service) (FnUnregister, error) {
	if instance == nil {
		return nil, ErrNoDefaultInstance
	}
	return instance.Register(s)
}

// Unregister unregister a service
//
// Call SetDefault before use
func Unregister(s Service) error {
	if instance == nil {
		return ErrNoDefaultInstance
	}
	return instance.Unregister(s)
}

func SetServiceStatus(s Service, status Status) error {
	if instance == nil {
		return ErrNoDefaultInstance
	}
	return instance.SetServiceStatus(s, status)
}

func GetRegisteredServices() ([]Service, error) {
	if instance == nil {
		return nil, ErrNoDefaultInstance
	}
	return instance.GetRegisteredServices(), nil
}

func (r *reg) GetService(name string, f ...Filter) (*Service, error) {
	services, err := r.getinternalService(name, &getServicesOptions{filters: f})
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, ErrNotFound
	}
	//return first item
	return &services[0], nil

}

func chainFilters(pongs []*Pong, filters ...Filter) []Service {
	services := append([]*Pong{}, pongs...)

	for _, f := range filters {
		if f != nil {
			services = f(services)
		}
	}

	res := make([]Service, len(services))
	for i, p := range services {
		res[i] = p.Service
	}
	return res
}

// observerGetOrCreate get observers entry if not exist create empty observer pointer
func (r *reg) observerGetOrCreate(key string) (o *observe, alreadyExist bool) {
	r.observersMu.Lock()
	defer r.observersMu.Unlock()
	if o, alreadyExist = r._observers[key]; !alreadyExist {
		o = &observe{}
		r._observers[key] = o
	}
	return
}

func newGetServicesOptions(options []func(*getServicesOptions)) *getServicesOptions {
	opts := &getServicesOptions{}
	for _, o := range options {
		o(opts)
	}
	return opts
}

// getinternalService retrieves a list of services by name, applying optional filters.
// If the service is already registered, it returns the filtered services immediately.
// If the service is not registered, it observes the service and waits for it to be registered,
// applying filters and returning the service if it matches the filters.
//
// Parameters:
// - name: The name of the service to retrieve.
// - opts: Options for retrieving the service, including filters and timeout settings.
//
// Returns:
// - services: A list of services that match the specified name and filters.
// - err: An error if the service is not found or if there is an issue with the retrieval process.
func (r *reg) getinternalService(name string, opts *getServicesOptions) (services []Service, err error) {
	filters := opts.filters
	if filters == nil {
		filters = r.opts.filters
	}
	//service is already registered
	r.log.Info("GetService ", name)
	if res := r.ser.GetServices(name); len(res) > 0 {
		if len(filters) > 0 {
			//if filters apply filters
			//=> if filters return empty response continue => send ping for
			if filtered := chainFilters(res, filters...); len(filtered) > 0 {
				return filtered, nil
			}
		} else {
			return chainFilters(res), nil
		}

	}

	obs, alreadyExist := r.observerGetOrCreate(name)
	if opts.nowait && alreadyExist {
		return nil, ErrNotFound
	}

	//service not yet registered
	//register invoke r.Observe(service) with a callback containing a channel
	//the callback apply filters on service and write in the channel when a service is ok with the filters
	ch := make(chan *Service, 1)

	obs.callback = func(p *Pong) {
		if filtered := chainFilters([]*Pong{p}, filters...); len(filtered) > 0 {
			obs.callback = nil
			ch <- &p.Service
		}
	}

	if !alreadyExist {
		r.Observe(name)
	}

	if err = r.opts.pubsub.Pub(r.buildMessage("ping", name), nil); err == nil {
		//create timeout if no service available
		var serviceFound *Service
		tk := time.NewTimer(timeout(opts.ctx, r.opts.timeout))
		select {
		case <-tk.C:
			break
		case serviceFound = <-ch:
			break
		}
		tk.Stop()
		if serviceFound != nil {
			services = []Service{*serviceFound}
		} else {
			err = ErrNotFound
		}
	}
	return
}
func timeout(ctx context.Context, timeout time.Duration) time.Duration {
	if ctx == nil {
		return timeout
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return timeout
	}
	return time.Until(deadline)
}

func (r *reg) GetServices(name string, opts ...func(*getServicesOptions)) ([]Service, error) {
	return r.getinternalService(name, newGetServicesOptions(opts))
}

func (r *reg) subregister(msg *pubsub.PubsubMsg) {
	var p *Pong
	err := json.Unmarshal(msg.Data, &p)
	if err != nil {
		r.log.Errorf("unmarshal error when sub to register: %v data: %s", err, string(msg.Data))
		return
	}
	r.log.Info("rcv ", p.Name, " kv ", p.KV)
	for _, f := range r.opts.observeFilters {
		if !f(p) {
			return
		}
	}
	o, _ := r.observerGetOrCreate(p.Name)
	if o.callback != nil {
		o.callback(p)
	}

	var alreadyExist bool
	if p, alreadyExist = r.ser.LoadOrStore(p); !alreadyExist && r.opts.observerEvent != nil {
		r.opts.observerEvent(p.Service, EventRegister)
	}
	if p.Timestamps != nil {
		d := int(float32(p.Timestamps.Duration) * r.opts.dueDurationFactor)
		registered := p.Timestamps.Registered * int64(time.Millisecond)
		p.dueTime = time.Unix(0, registered).Add(time.Duration(d) * time.Millisecond)
	}
	r.log.Debugf("append %s ", p.Service)

}

func (r *reg) subunregister(msg *pubsub.PubsubMsg) {
	var s Service
	err := json.Unmarshal(msg.Data, &s)
	if err != nil {
		r.log.Errorf("unmarshal error when sub to register: %s", err)
		return
	}
	p := &Pong{Service: s}
	for _, f := range r.opts.observeFilters {
		if !f(p) {
			return
		}
	}
	if r.opts.observerEvent != nil {
		r.opts.observerEvent(s, EventUnregister)
	}
	r.ser.DeleteByName(s.Name + s.Address)
	r.log.Debugf("Unregister service %s/%s", s.Name, s.Address)
}

// adding subscription
// register to service
// unregister to service
func (r *reg) Observe(service string) (err error) {
	r.observerGetOrCreate(service)

	var s pubsub.Subscription
	s, err = r.opts.pubsub.Sub(r.buildMessage("register", service), r.subregister)
	if err == nil {
		r.subscriptions = append(r.subscriptions, s)
		s, err = r.opts.pubsub.Sub(r.buildMessage("unregister", service), r.subunregister)
		if err == nil {
			r.subscriptions = append(r.subscriptions, s)
		}
	}
	return
}

// Close unregister to all subscriptions.
// Clear local cache.
// Stop go routine if exist.
// TODO
func (r *reg) Close() (err error) {
	r.chStopChannelRegisteredServices <- true
	r.registeredServicesMap.Range(func(k, v interface{}) bool {
		r.Unregister(v.(*Pong).Service)
		return true
	})

	for _, s := range r.subscriptions {
		s.Unsub()
	}
	r.subscriptions = r.subscriptions[0:0]
	instance = nil
	r.log.Debug("Close registry done")
	return
}
