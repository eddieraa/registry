package nats

import (
	"github.com/eddieraa/registry"
	"github.com/nats-io/nats.go"
)

type pubsub struct {
	c *nats.Conn
}
type subscription struct {
	s *nats.Subscription
}

//NewPub return NATS Pubsub
func NewPub(c *nats.Conn) registry.Pubsub {
	pubsub := &pubsub{
		c: c,
	}
	return pubsub
}
func (pb *pubsub) Sub(topic string, f func(m *registry.PubsubMsg)) (registry.Subscription, error) {
	subscript, err := pb.c.Subscribe(topic, func(m *nats.Msg) {
		f(&registry.PubsubMsg{Subject: m.Subject, Data: m.Data})
	})
	s := &subscription{s: subscript}
	return s, err
}
func (pb *pubsub) Pub(topic string, data []byte) error {
	return pb.c.Publish(topic, data)
}

func (s *subscription) Unsub() error {
	return s.s.Unsubscribe()
}
func (s *subscription) Subject() string {
	return s.s.Subject
}

//Nats initialyse service registry with nats connection
func Nats(conn *nats.Conn) registry.Option {
	return registry.WithPubsub(NewPub(conn))
}
