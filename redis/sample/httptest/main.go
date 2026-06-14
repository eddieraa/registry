package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/eddieraa/registry"
	"github.com/eddieraa/registry/redis"
)

func main() {
	var log = slog.Default()

	r, err := registry.NewRegistry(redis.NewRedisClient(""))
	if err != nil {
		log.Error("unable to connect to redis ", "error", err.Error())
		os.Exit(1)
	}
	_, err = r.Register(registry.Service{Address: "localhost:5435", Name: "httptest"})

	//Intercep CTRL-C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Info("Stop")

}
