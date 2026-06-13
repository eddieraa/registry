package registry

import (
	"log/slog"
	"os"
)

type Logger interface {
	Debug(msg string, args ...any)

	Error(msg string, args ...any)

	Info(msg string, args ...any)

	Warn(msg string, args ...any)
}

func initLogger(o *Options) {
	if o.logger == nil {
		o.logger = slog.New(slog.NewTextHandler(
			os.Stdout, &slog.HandlerOptions{
				Level:     o.loglevel,
				AddSource: true,
			},
		))
	}
}
