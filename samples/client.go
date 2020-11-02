package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eddieraa/registry"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

func main() {
	var serviceName string
	flag.StringVar(&serviceName, "service-name", "httptest", "http service name")
	var natsURL string
	flag.StringVar(&natsURL, "nats-url", "localhost:4222", "NATS server URL ")
	flag.Parse()

	conn, err := nats.Connect(natsURL)
	if err != nil {
		logrus.Fatal("could not connect to nats ", err)
	}
	reg, err := registry.Connect(registry.Nats(conn))
	if err != nil {
		logrus.Fatal("could not connect to nats ", err)
	}
	service, err := reg.GetService(serviceName, registry.LoadBalanceFilter())
	if err != nil {
		logrus.Fatal("Could not get service ", err)
	}

	rep, err := http.Get(fmt.Sprintf("http://%s/", service.Address))
	if err != nil {
		logrus.Fatal("Could net request url ", err)
	}

	out, _ := ioutil.ReadAll(rep.Body)
	logrus.Info("Read ", string(out))
}
