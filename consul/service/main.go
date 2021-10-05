package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/eddieraa/registry"
	"github.com/eddieraa/registry/consul"
	regnats "github.com/eddieraa/registry/nats"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)
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

	registry.NewRegistry(regnats.Nats(conn))
	consul.HandleServices()
	log.Fatal(http.ListenAndServe(bindAddress, nil))
}
