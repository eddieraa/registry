package main

import (
	"os"

	"github.com/eddieraa/registry"
	"github.com/eddieraa/registry/redis"
)

func main() {
	var log = registry.NewDefaulLogger()
	log.SetLevel(registry.DebugLevel)
	r, err := registry.NewRegistry(redis.NewRedisClient(""))
	if err != nil {
		log.Error("Unable to create registry ", err)
		os.Exit(1)
	}

	s, err := r.GetService("httptest")
	if err != nil {
		log.Error("Unable to get service ", err)
		os.Exit(1)
	}
	log.Error("Found ", s.Name, " ", s.Address)
}
