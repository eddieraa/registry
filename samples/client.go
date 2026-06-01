package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"

	"github.com/eddieraa/registry"
	pb "github.com/eddieraa/registry/nats"
	"github.com/nats-io/nats.go"
)

func main() {
	var serviceName string
	flag.StringVar(&serviceName, "service-name", "httptest", "http service name")
	var natsURL string
	flag.StringVar(&natsURL, "nats-url", "localhost:4222", "NATS server URL ")
	var loadBalance bool
	flag.BoolVar(&loadBalance, "load-balance", false, "Activate load balancing")

	//parse
	flag.Parse()

	var log = slog.Default()

	conn, err := nats.Connect(natsURL)
	if err != nil {
		log.Error("could not connect to nats ", "error", err.Error())
		os.Exit(1)
	}
	reg, err := registry.SetDefault(pb.Nats(conn), registry.AddFilter(registry.LoadBalanceFilter()))
	if err != nil {
		log.Error("could not connect to nats ", "error", err.Error())
		os.Exit(1)
	}

	for i := 0; i < 10; i++ {
		service, err := reg.GetService(serviceName)
		if err != nil {
			log.Error("Could not get service %s: %v", "service", serviceName, "error", err.Error())
		}

		rep, err := http.Get(fmt.Sprintf("http://%s/httptest", service.Address))
		if err != nil {
			log.Error("Could not request url ", "error", err.Error())
		}
		out, _ := ioutil.ReadAll(rep.Body)
		log.Info("Read ", string(out))
	}
	reg.Close()

}
