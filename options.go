package registry

import (
	"time"

	"github.com/eddieraa/registry/pubsub"
)

//Option option func
type Option func(opts *Options)

//Options all configurable option
type Options struct {
	timeout           time.Duration
	registerInterval  time.Duration
	checkDueTime      time.Duration
	pubsub            pubsub.Pubsub
	mainTopic         string
	filters           []Filter
	observeFilters    []ObserveFilter
	dueDurationFactor float32
	observerEvent     ObserverEvent
	hostname          string
}

var (
	//DefaultTimeout timeout for GetServices
	DefaultTimeout = 100 * time.Millisecond
	//DefaultRegisterInterval time between 2 registration
	DefaultRegisterInterval = 20 * time.Second
	//DefaultCheckDueInterval time between 2 checks
	DefaultCheckDueInterval = 200 * time.Millisecond
	//DefaultMainTopic default message base. All topics will start with this message
	DefaultMainTopic = "registry"
	//DefaultDueDurationFactor service expire when currentTime > lastRegisteredTime + registerInternal * dueDrationFactor
	DefaultDueDurationFactor = float32(1.5)
)

/*
//SetFlags set go flags.
// Call this func if you want to override default parameters with command line argument
//func SetFlags() {
//flag.IntVar(&DefaultTimeout, "registry-timeout", 100,"s")

//}
*/

func newOptions(opts ...Option) Options {
	options := Options{
		timeout:           DefaultTimeout,
		registerInterval:  DefaultRegisterInterval,
		checkDueTime:      DefaultCheckDueInterval,
		mainTopic:         DefaultMainTopic,
		dueDurationFactor: DefaultDueDurationFactor,
		filters:           make([]Filter, 0),
		observeFilters:    make([]ObserveFilter, 0),
	}
	options.hostname = hostname()
	for _, o := range opts {
		o(&options)
	}
	return options
}

//WithPubsub initialyse service registry with nats connection
func WithPubsub(pb pubsub.Pubsub) Option {
	return func(opts *Options) {
		opts.pubsub = pb
	}
}

//Timeout define timeout
func Timeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.timeout = timeout
	}
}

//RegisterInterval time between 2 register publish
func RegisterInterval(duration time.Duration) Option {
	return func(opts *Options) {
		opts.registerInterval = duration
	}
}

//MainTopic all topic will start with topic. Usefull in multi-tenancy
func MainTopic(topic string) Option {
	return func(opts *Options) {
		opts.mainTopic = topic
	}
}

//AddFilter add filter
func AddFilter(f Filter) Option {
	return func(opts *Options) {
		opts.filters = append(opts.filters, f)
	}
}

//AddObserveFilter adding filter
func AddObserveFilter(f ObserveFilter) Option {
	return func(opts *Options) {
		opts.observeFilters = append(opts.observeFilters, f)
	}
}

//SetObserverEvent set handler for Observer Event
func SetObserverEvent(ev ObserverEvent) Option {
	return func(opts *Options) {
		opts.observerEvent = ev
	}
}
