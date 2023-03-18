package test

import (
	"fmt"

	"github.com/eddieraa/registry/pubsub"
)

var count int

type cli struct {
	i      int
	server *fakeServer
	cbPub  *func(string, []byte) ([]byte, error)
	cbSub  *func(string, *pubsub.PubsubMsg) *pubsub.PubsubMsg
}

func newCli(s *fakeServer) pubsub.Pubsub {
	c := &cli{i: count, server: s}
	count++
	return c
}

func (c *cli) CallbackSub(f func(string, *pubsub.PubsubMsg) *pubsub.PubsubMsg) {
	if f == nil {
		c.cbSub = nil
	} else {
		c.cbSub = &f
	}
}

func (c *cli) CallbackPub(f func(string, []byte) ([]byte, error)) {
	if f == nil {
		c.cbPub = nil
	} else {
		c.cbPub = &f
	}
}

func (c *cli) Pub(topic string, data []byte) (err error) {
	if c.cbPub != nil {
		f := *c.cbPub
		if data, err = f(topic, data); err != nil {
			return
		}
	}
	err = c.server.SendMessage(&pubsub.PubsubMsg{Subject: topic, Data: data})
	return
}

func (c *cli) Stop() {
	c.server.Pause()
}

func (c *cli) Sub(topic string, f func(m *pubsub.PubsubMsg)) (res pubsub.Subscription, err error) {
	if c.cbSub != nil {
		fct := *c.cbSub
		newf := func(m *pubsub.PubsubMsg) {
			newm := fct(topic, m)
			f(newm)
		}
		return c.server.Sub(c, topic, newf)
	}
	return c.server.Sub(c, topic, f)
}
func (c *cli) String() string {
	return fmt.Sprintf("%d", c.i)
}

// Debug debug interface
type Debug interface {
	CallbackSub(func(string, *pubsub.PubsubMsg) *pubsub.PubsubMsg)
	CallbackPub(func(string, []byte) ([]byte, error))
}
