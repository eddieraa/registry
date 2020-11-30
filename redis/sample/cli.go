package main

import (
	"github.com/eddieraa/registry"
	"github.com/eddieraa/registry/redis"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	r, err := registry.NewRegistry(redis.NewRedisClient(""))
	if err != nil {
		logrus.Fatal("Unable to create registry ", err)
	}

	s, err := r.GetService("httptest")
	if err != nil {
		logrus.Fatal("Unable to get service ", err)
	}
	logrus.Print("Found ", s.Name, " ", s.Address)
}
