package test

import (
	"sync"

	"github.com/eddieraa/registry"
)

type fakeServer struct {
	subsc sync.Map
}

func (s *fakeServer) add(topic string, c *cli, f func(m *registry.PubsubMsg)) {
	var clis map[cli]func(m *registry.PubsubMsg)
	v, ok := s.subsc.Load(topic)
	if ok {
		clis = v.(map[cli]func(m *registry.PubsubMsg))
	} else {
		clis = make(map[cli]func(m *registry.PubsubMsg))
		s.subsc.Store(topic, clis)
	}
	clis[*c] = f
}
func (s *fakeServer) get(topic string) map[cli]func(*registry.PubsubMsg) {
	var res map[cli]func(*registry.PubsubMsg)
	if v, ok := s.subsc.Load(topic); ok {
		res = v.(map[cli]func(*registry.PubsubMsg))
	}
	return res
}

func (s *fakeServer) sizeTopic(topic string) int {
	topics := s.get(topic)
	if topics == nil {
		return 0
	}
	return len(topics)
}

var server fakeServer

type cli struct {
	i int
}
type subscription struct {
	subject string
	unsub   func()
}

func (s *subscription) Subject() string {
	return s.subject
}
func (s *subscription) Unsub() (err error) {
	s.unsub()
	return
}

func init() {

}
func newFakePubsub() registry.Pubsub {
	c := &cli{}

	return c
}
func (c *cli) Pub(topic string, data []byte) error {
	if clis, ok := server.subsc.Load(topic); ok {
		for _, f := range clis.(map[cli]func(m *registry.PubsubMsg)) {
			f(&registry.PubsubMsg{Subject: topic, Data: data})
		}
	}
	return nil
}

func (c *cli) Sub(topic string, f func(m *registry.PubsubMsg)) (registry.Subscription, error) {
	server.add(topic, c, f)
	res := &subscription{subject: topic, unsub: func() {
		if s, ok := server.subsc.Load(topic); ok {
			delete(s.(map[cli]bool), *c)
		}
	}}
	return res, nil
}
