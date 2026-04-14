package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/eddieraa/registry"
	"github.com/eddieraa/registry/redis"
)

func main() {
	var log = registry.NewDefaulLogger()
	log.SetLevel(registry.DebugLevel)
	r, err := registry.NewRegistry(redis.NewRedisClient(""))
	if err != nil {
		log.Fatal("unable to connect to redis ", err)
	}
	_, err = r.Register(registry.Service{Address: "localhost:5435", Name: "httptest"})

	//Intercep CTRL-C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Info("Stop")

}
