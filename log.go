package registry

import "log/slog"

type Logger interface {
	Debug(msg string, args ...any)

	Error(msg string, args ...any)

	Info(msg string, args ...any)

	Warn(msg string, args ...any)
}

func initLogger(o *Options) {
	if o.logger == nil {
		o.logger = slog.New(slog.Default().Handler())

	}
	// Set log level if the logger supports it
	if levelSetter, ok := o.logger.(interface {
		SetLevel(slog.Level)
	}); ok {
		levelSetter.SetLevel(o.loglevel)
	}
}
