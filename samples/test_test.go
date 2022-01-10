package main

import (
	"testing"

	"github.com/eddieraa/registry"
	pb "github.com/eddieraa/registry/nats"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

func Test1(t *testing.T) {
	conn, err := nats.Connect("nats://localhost")
	if err != nil {
		logrus.Fatal("could not connect to nats ", err)
	}
	reg, err := registry.SetDefault(pb.Nats(conn))
	if err != nil {
		logrus.Fatal("could not connect to nats ", err)
	}
	//reg.Observe("java-test")
	s, err := reg.GetService("java-test")
	if err != nil {
		logrus.Fatal("Could not get service java-test ", err)
	}
	logrus.Info("YEAP ", s.Address)

}
