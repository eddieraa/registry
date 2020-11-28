package test

import (
	"strings"
	"sync"

	"github.com/eddieraa/registry"
)

type fakeServer struct {
	fm       *fakemap
	fmlike   *fakemap
	cachesub map[string][]func(*registry.PubsubMsg)
	mu       sync.Mutex
}

var server fakeServer

func init() {
	server.fm = newFakemap()
	server.fmlike = newFakemap()
}

func (s *fakeServer) getCachesub() map[string][]func(*registry.PubsubMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cachesub == nil {
		s.cachesub = make(map[string][]func(*registry.PubsubMsg))
	}
	return s.cachesub

}

func (s *fakeServer) rebuildCache(c *cli, topic string, f func(m *registry.PubsubMsg)) {
	newCache := make(map[string][]func(*registry.PubsubMsg))
	s.fm.Range(func(topic string, msgs []func(*registry.PubsubMsg)) {
		newCache[topic] = msgs
	})

	s.fmlike.Range(func(topic string, msgs []func(*registry.PubsubMsg)) {
		//todo
	})

}

func (s *fakeServer) Sub(c *cli, topic string, f func(m *registry.PubsubMsg)) (res registry.Subscription) {
	if strings.HasSuffix(topic, "*") {
		server.fmlike.add(topic, c, f)
	} else {
		server.fm.add(topic, c, f)
	}
	return
}

func (s *fakeServer) SendMessage(mes *registry.PubsubMsg) {
	m := s.getCachesub()
	if subs, ok := m[mes.Subject]; ok {
		for _, v := range subs {
			v(mes)
		}
	}

}

//NewPubSub return new Pubsub instance
func NewPubSub() registry.Pubsub {
	return newCli(&server)
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
