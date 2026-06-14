package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/eddieraa/registry"
	regnats "github.com/eddieraa/registry/nats"
	consul "github.com/eddieraa/registry/rest"
	"github.com/nats-io/nats.go"
)

func main() {
	var log = slog.Default()

	var natsURL string
	flag.StringVar(&natsURL, "nats-url", "localhost:4222", "NATS server URL ")

	var bindAddress string
	flag.StringVar(&bindAddress, "port", ":5454", "address to bind")

	var services string
	flag.StringVar(&services, "services", "", "services ex: --services=service1,service2,service3")

	flag.Parse()

	conn, err := nats.Connect(natsURL)
	if err != nil {
		panic(fmt.Sprint("Could not connect to nats (", natsURL, "): ", err))
	}

	r, err := registry.NewRegistry(regnats.Nats(conn))
	if err != nil {
		log.Error("unable to create registry ", "error", err.Error())
		os.Exit(1)
	}
	r.Observe("*")
	r.Register(registry.Service{Name: consul.REGISTRY_NAME, Address: bindAddress})
	consul.HandleServices(r)
	if err := http.ListenAndServe(bindAddress, nil); err != nil {
		log.Error("unable to start server ", "error", err.Error())
		os.Exit(1)
	}
}
