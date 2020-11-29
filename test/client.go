package test

import (
	"fmt"

	"github.com/eddieraa/registry/pubsub"
)

var count int

type cli struct {
	i      int
	server *fakeServer
}

func newCli(s *fakeServer) pubsub.Pubsub {
	c := &cli{i: count, server: s}
	count++
	return c
}

func (c *cli) Pub(topic string, data []byte) error {
	c.server.SendMessage(&pubsub.PubsubMsg{Subject: topic, Data: data})
	return nil
}

func (c *cli) Sub(topic string, f func(m *pubsub.PubsubMsg)) (res pubsub.Subscription, err error) {
	return c.server.Sub(c, topic, f), nil
}
func (c *cli) String() string {
	return fmt.Sprintf("%d", c.i)
}
