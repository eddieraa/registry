package redis

import (
	"context"

	"github.com/eddieraa/registry"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type pubsub struct {
	rb            *redis.Client
	pb            *redis.PubSub
	ctx           context.Context
	subscriptions map[string]*subscription
	chstop        chan bool
	connManaged   bool
}

type subscription struct {
	channel string
	unsub   func() error
	sub     func(m *registry.PubsubMsg)
}

//NewRedisClient return Options with redis connection
func NewRedisClient(addr string) registry.Option {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	pb := NewPub(rdb)
	pb.(*pubsub).connManaged = true
	return registry.WithPubsub(pb)
}

//WithRedis return Option with redis client
func WithRedis(rdb *redis.Client) registry.Option {
	return registry.WithPubsub(NewPub(rdb))
}

//NewPub return Redis Pubsub
func NewPub(rb *redis.Client) registry.Pubsub {
	ctx := context.Background()
	res := &pubsub{
		rb:            rb,
		ctx:           ctx,
		pb:            rb.Subscribe(ctx),
		subscriptions: make(map[string]*subscription),
		chstop:        make(chan bool),
	}
	go res.subscribing()
	return res
}

func (p *pubsub) subscribing() {
	ch := p.pb.Channel()
stop:
	for {
		select {
		case <-p.chstop:
			logrus.Info("Stop requested")
			break stop
		case m := <-ch:
			if s, ok := p.subscriptions[m.Channel]; ok {
				s.sub(&registry.PubsubMsg{Subject: m.Channel, Data: []byte(m.Payload)})
			}
		}
	}
}

func (p *pubsub) Sub(topic string, f func(m *registry.PubsubMsg)) (registry.Subscription, error) {
	logrus.Debug("sub ", topic)
	err := p.pb.Subscribe(p.ctx, topic)
	if err != nil {
		return nil, err
	}
	subscription := &subscription{
		unsub:   func() error { return p.pb.Unsubscribe(p.ctx, topic) },
		channel: topic,
		sub:     f,
	}
	p.subscriptions[topic] = subscription
	return subscription, nil
}

func (p *pubsub) Pub(topic string, data []byte) error {
	logrus.Debug("pub ", topic)
	publish := p.rb.Publish(p.ctx, topic, data)
	return publish.Err()
}

func (p *pubsub) Stop() {
	p.chstop <- true
	if p.connManaged {
		p.rb.Close()
	}
}

func (s *subscription) Unsub() error {
	logrus.Debug("unsub ", s.channel)
	return s.unsub()
}
func (s *subscription) Subject() string {
	return s.channel
}
