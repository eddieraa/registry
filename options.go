package registry

import (
	"time"

	"github.com/nats-io/nats.go"
)

//Option option func
type Option func(opts *Options)

//Options all configurable option
type Options struct {
	timeout          time.Duration
	registerInterval time.Duration
	natsConn         *nats.Conn
}

var (
	//DefaultTimeout timeout on for GetServices func
	DefaultTimeout = 100 * time.Millisecond
	//DefaultRegisterInterval time between 2 registration
	DefaultRegisterInterval = 20 * time.Second
)

func newOptions(opts ...Option) Options {
	options := Options{
		timeout:          DefaultTimeout,
		registerInterval: DefaultRegisterInterval,
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
