package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eddieraa/registry"
	pb "github.com/eddieraa/registry/nats"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
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

	conn, err := nats.Connect(natsURL)
	if err != nil {
		logrus.Fatal("could not connect to nats ", err)
	}
	reg, err := registry.SetDefault(pb.Nats(conn), registry.AddFilter(registry.LoadBalanceFilter()))
	if err != nil {
		logrus.Fatal("could not connect to nats ", err)
	}

	for i := 0; i < 10; i++ {
		service, err := reg.GetService(serviceName)
		if err != nil {
			logrus.Fatalf("Could not get service %s: %v", serviceName, err)
		}

		rep, err := http.Get(fmt.Sprintf("http://%s/httptest", service.Address))
		if err != nil {
			logrus.Fatal("Could net request url ", err)
		}
		out, _ := ioutil.ReadAll(rep.Body)
		logrus.Info("Read ", string(out))
	}
	reg.Close()

}
