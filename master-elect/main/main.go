package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eddieraa/registry"
	masterelect "github.com/eddieraa/registry/master-elect"
	nr "github.com/eddieraa/registry/nats"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	flag.Bool("list", false, "list registered services and exit")
	flag.Parse()

	slog.Info("starting")
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if flag.Arg(0) == "list" {
		r, err := registry.NewRegistry(nr.Nats(conn), registry.WithLoglevel(logrus.InfoLevel))
		if err != nil {
			panic(err)
		}
		defer r.Close()
		r.Observe("myService")
		// sleep a bit to be sure to receive the observe event
		for {
			services, err := r.GetServices("myService")
			if err != nil {
				logrus.Error("failed to get services", "error", err)
			}
			println("")
			for _, s := range services {
				//logrus.Info("service ", s.Address, " with id ", s.KV["id"], " master(", s.KV["master"], ")")
				logrus.Infof("service %s with address %s master(%s)", s.KV["id"], s.Address, s.KV["master"])
			}
			time.Sleep(5 * time.Second)
		}

	}
	r, err := registry.NewRegistry(nr.Nats(conn), registry.WithLoglevel(logrus.DebugLevel))
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
