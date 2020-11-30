package nats

import (
	"github.com/eddieraa/registry"
	"github.com/eddieraa/registry/pubsub"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type pb struct {
	c *nats.Conn
}
type subscription struct {
	s *nats.Subscription
}

//NewPub return NATS Pubsub
func NewPub(c *nats.Conn) pubsub.Pubsub {
	pb := &pb{
		c: c,
	}
	return pb
}
func (pb *pb) Sub(topic string, f func(m *pubsub.PubsubMsg)) (pubsub.Subscription, error) {
	logrus.Debug("subscribe to: ", topic)
	subscript, err := pb.c.Subscribe(topic, func(m *nats.Msg) {
		f(&pubsub.PubsubMsg{Subject: m.Subject, Data: m.Data})
	})
	s := &subscription{s: subscript}
	return s, err
}
func (pb *pb) Pub(topic string, data []byte) error {
	logrus.Debug("publish: ", topic)
	return pb.c.Publish(topic, data)
}

func (pb *pb) Stop() {

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
