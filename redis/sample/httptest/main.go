package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/eddieraa/registry"
	"github.com/eddieraa/registry/redis"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	r, err := registry.Connect(redis.NewRedisClient(""))
	if err != nil {
		logrus.Fatal("unable to connect to redis ", err)
	}
	_, err = r.Register(registry.Service{Address: "localhost:5435", Name: "httptest"})

	//Intercep CTRL-C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logrus.Info("Stop")

}
