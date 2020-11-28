package test

import (
	"fmt"

	"github.com/eddieraa/registry"
)

var count int

type cli struct {
	i      int
	server *fakeServer
}

func newCli(s *fakeServer) registry.Pubsub {
	c := &cli{i: count, server: s}
	count++
	return c
}

func (c *cli) Pub(topic string, data []byte) error {
	c.server.SendMessage(&registry.PubsubMsg{Subject: topic, Data: data})
	return nil
}

func (c *cli) Sub(topic string, f func(m *registry.PubsubMsg)) (res registry.Subscription, err error) {
	return c.server.Sub(c, topic, f), nil
}
func (c *cli) String() string {
	return fmt.Sprintf("%d", c.i)
}
