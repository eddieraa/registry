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

var log = logrus.New()

// SetDefault with a nats connection
func SetDefault(c *nats.Conn, opts ...registry.Option) (r registry.Registry, err error) {
	options := []registry.Option{Nats(c)}
	if opts != nil {
		options = append(options, opts...)
	}

	return registry.SetDefault(options...)
}

// NewPub return NATS Pubsub
func NewPub(c *nats.Conn) pubsub.Pubsub {
	pb := &pb{
		c: c,
	}
	return pb
}
func (pb *pb) Sub(topic string, f func(m *pubsub.PubsubMsg)) (pubsub.Subscription, error) {
	log.Debug("subscribe to: ", topic)
	subscript, err := pb.c.Subscribe(topic, func(m *nats.Msg) {
		f(&pubsub.PubsubMsg{Subject: m.Subject, Data: m.Data})
	})
	s := &subscription{s: subscript}
	return s, err
}
func (pb *pb) Pub(topic string, data []byte) error {
	log.Debug("publish: ", topic)
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

// Nats initialyse service registry with nats connection
func Nats(conn *nats.Conn) registry.Option {
	return registry.WithPubsub(NewPub(conn))
}

// SetLogLevel log level
func SetLogLevel(level logrus.Level) {
	log.SetLevel(level)
}
