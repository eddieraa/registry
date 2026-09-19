package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/eddieraa/registry"
	masterelect "github.com/eddieraa/registry/master-elect"
	nr "github.com/eddieraa/registry/nats"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("15:04:05"))
			}
			return a
		},
	}))
	slog.SetDefault(logger)
	flag.Bool("list", false, "list registered services and exit")
	flag.Parse()

	slog.Info("starting")
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if flag.Arg(0) == "list" {
		r, err := registry.NewRegistry(nr.Nats(conn), registry.WithLoglevel(logrus.InfoLevel), registry.WithLogger(logger))
		if err != nil {
			panic(err)
		}
		defer r.Close()
		r.Observe("myService")
		// sleep a bit to be sure to receive the observe event
		r.AddObserverEvent(func(s registry.Service, ev registry.Event) {
			services, err := r.GetServices("myService")
			if err != nil {
				slog.Error("failed to get services", "error", err)
			}
			slog.Info("=======> new", "event", ev, "service ", s.Address)
			for _, s := range services {
				slog.Info("service", "id", s.KV["id"], "address", s.Address, "master", s.KV["elect-master"])
			}
		})
		waitForCtrlCSignal()
		return

	}
	r, err := registry.NewRegistry(nr.Nats(conn), registry.WithLoglevel(logrus.ErrorLevel), registry.WithLogger(logger))
	if err != nil {
		panic(err)
	}
	defer r.Close()
	fnUnregister, err := register(r)
	if err != nil {
		panic(err)
	}

	defer func() {
		slog.Info("attempt to close")
		fnUnregister()
	}()

	err = masterelect.New(r, "myService")
	if err != nil {
		panic(err)
	}

	// wait for ctrl+c
	waitForCtrlCSignal()
}

func register(r registry.Registry) (func(), error) {
	address := flag.Arg(0)
	if address == "" {
		address = fmt.Sprintf(":%d", os.Getpid())
	}
	fnUnregister, err := r.Register(registry.Service{
		Name:    "myService",
		Address: address,
		KV: map[string]string{
			"myservice": fmt.Sprintf("%d", os.Getpid()),
			"elect-id":  address,
		},
	})
	if err != nil {
		return func() { r.Close() }, err
	}
	return func() {
		slog.Info("unregistering service")
		fnUnregister()

	}, nil
}

func waitForCtrlCSignal() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,    // SIGINT (Ctrl+C)
		syscall.SIGTERM, // termination signal
	)

	<-ctx.Done()
	stop()
	slog.Info("ctrl+c received, shutting down")
}
