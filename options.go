package registry

import "time"

//Option option func
type Option func(opts *Options)

//Options all configurable option
type Options struct {
	timeout time.Duration
}

var (
	//DefaultTimeout timeout on for GetServices func
	DefaultTimeout = 100 * time.Millisecond
)

func newOptions(opts ...Option) Options {
	options := Options{
		timeout: DefaultTimeout,
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

//Timeout define timeout
func Timeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.timeout = timeout
	}
}
