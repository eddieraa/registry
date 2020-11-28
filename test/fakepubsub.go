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
	fakecli  registry.Pubsub
}

var server fakeServer

func init() {
	server.fm = newFakemap()
	server.fmlike = newFakemap()
	server.fakecli = newCli(&server)
}

func (s *fakeServer) getCachesub() map[string][]func(*registry.PubsubMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cachesub == nil {
		s.cachesub = make(map[string][]func(*registry.PubsubMsg))
	}
	return s.cachesub

}

func (s *fakeServer) rebuildAllCache() {
	newCache := make(map[string][]func(*registry.PubsubMsg))
	s.fm.Range(func(topic string, msgs []func(*registry.PubsubMsg)) {
		newCache[topic] = msgs
	})

	s.fmlike.Range(func(topic string, msgs []func(*registry.PubsubMsg)) {
		prefix := topic[0 : len(topic)-1]
		topics := s.fm.FindTopicsWithPrefix(prefix)
		for _, t := range topics {
			var cachemsgs []func(*registry.PubsubMsg)
			var ok bool
			if cachemsgs, ok = newCache[t]; !ok {
				cachemsgs = msgs
			} else {
				cachemsgs = append(cachemsgs, msgs...)
			}
			newCache[t] = cachemsgs
		}
	})
	s.mu.Lock()
	s.cachesub = newCache
	s.mu.Unlock()

}

func (s *fakeServer) Unsub(c *cli, topic string) {
	var fm *fakemap
	if strings.HasSuffix(topic, "*") {
		fm = s.fmlike
	} else {
		fm = s.fm
	}
	if fm.Remove(topic, c) {
		s.rebuildAllCache()
	}
}

func (s *fakeServer) Sub(c *cli, topic string, f func(m *registry.PubsubMsg)) (res registry.Subscription) {
	if strings.HasSuffix(topic, "*") {
		server.fmlike.add(topic, c, f)
	} else {
		server.fm.add(topic, c, f)
	}
	s.rebuildAllCache()

	res = &subscription{subject: topic, unsub: func() { s.Unsub(c, topic) }}
	return
}

func (s *fakeServer) SendMessage(mes *registry.PubsubMsg) {
	m := s.getCachesub()

	if subs, ok := m[mes.Subject]; ok {
		for _, v := range subs {
			v(mes)
		}
	} else {
		//subject is unknown
		s.fakecli.Sub(mes.Subject, fakesub)
		m = s.getCachesub()
		if subs, ok = m[mes.Subject]; ok {
			for _, v := range subs {
				v(mes)
			}
		}
	}

}

func fakesub(m *registry.PubsubMsg) {
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
