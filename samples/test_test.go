//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/eddieraa/registry"
	pb "github.com/eddieraa/registry/nats"
	"github.com/nats-io/nats.go"
)

func Test1(t *testing.T) {
	var log = registry.NewDefaulLogger()
	conn, err := nats.Connect("nats://localhost")
	if err != nil {
		log.Fatal("could not connect to nats ", err)
	}
	reg, err := registry.SetDefault(pb.Nats(conn))
	if err != nil {
		log.Fatal("could not connect to nats ", err)
	}
	//reg.Observe("java-test")
	s, err := reg.GetService("java-test")
	if err != nil {
		log.Fatal("Could not get service java-test ", err)
	}
	log.Info("YEAP ", s.Address)

}
