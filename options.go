package registry

import (
	"time"

	"github.com/nats-io/nats.go"
)

//Option option func
type Option func(opts *Options)

//Options all configurable option
type Options struct {
	timeout           time.Duration
	registerInterval  time.Duration
	checkDueTime      time.Duration
	natsConn          *nats.Conn
	mainTopic         string
	filters           []Filter
	dueDurationFactor float32
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

//SetFlags set go flags.
// Call this func if you want to override default parameters with command line argument
func SetFlags() {
	//flag.IntVar(&DefaultTimeout, "registry-timeout", 100,"s")

}

func newOptions(opts ...Option) Options {
	options := Options{
		timeout:           DefaultTimeout,
		registerInterval:  DefaultRegisterInterval,
		checkDueTime:      DefaultCheckDueInterval,
		mainTopic:         DefaultMainTopic,
		dueDurationFactor: DefaultDueDurationFactor,
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

//Nats initialyse service registry with nats connection
func Nats(conn *nats.Conn) Option {
	return func(opts *Options) {
		opts.natsConn = conn
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
		if opts.filters == nil {
			opts.filters = []Filter{}
		}
		opts.filters = append(opts.filters, f)
	}
}
