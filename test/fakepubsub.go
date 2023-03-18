package test

import (
	"errors"
	"strings"
	"sync"

	"github.com/eddieraa/registry/pubsub"
)

type fakeServer struct {
	fm       *fakemap
	fmlike   *fakemap
	cachesub map[string][]func(*pubsub.PubsubMsg)
	mu       sync.Mutex
	fakecli  pubsub.Pubsub
	pause    bool
}

// ErrPubsubPaused when pubsub is not available
var ErrPubsubPaused = errors.New("fakepubsub is paused")

// Server ability to Pause or Restart
type Server interface {
	Resume()
	Pause()
}

var server fakeServer

func init() {
	server.fm = newFakemap()
	server.fmlike = newFakemap()
	server.fakecli = newCli(&server)
	server.pause = false
}

func (s *fakeServer) getCachesub() map[string][]func(*pubsub.PubsubMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cachesub == nil {
		s.cachesub = make(map[string][]func(*pubsub.PubsubMsg))
	}
	return s.cachesub

}

func (s *fakeServer) Pause() {
	s.pause = true
}

func (s *fakeServer) Resume() {
	s.pause = false
}

func (s *fakeServer) rebuildAllCache() {
	newCache := make(map[string][]func(*pubsub.PubsubMsg))
	s.fm.Range(func(topic string, msgs []func(*pubsub.PubsubMsg)) {
		newCache[topic] = msgs
	})

	s.fmlike.Range(func(topic string, msgs []func(*pubsub.PubsubMsg)) {
		prefix := topic[0 : len(topic)-1]
		topics := s.fm.FindTopicsWithPrefix(prefix)
		for _, t := range topics {
			var cachemsgs []func(*pubsub.PubsubMsg)
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

func (s *fakeServer) Sub(c *cli, topic string, f func(m *pubsub.PubsubMsg)) (res pubsub.Subscription, err error) {
	if s.pause {
		err = ErrPubsubPaused
		return
	}
	if strings.HasSuffix(topic, "*") {
		server.fmlike.add(topic, c, f)
	} else {
		server.fm.add(topic, c, f)
	}
	s.rebuildAllCache()

	res = &subscription{subject: topic, unsub: func() { s.Unsub(c, topic) }}
	return
}

func (s *fakeServer) SendMessage(mes *pubsub.PubsubMsg) error {
	if s.pause {
		//don't send message
		return ErrPubsubPaused
	}
	m := s.getCachesub()

	if subs, ok := m[mes.Subject]; ok {
		for _, v := range subs {
			go v(mes)
		}
	} else {
		//subject is unknown
		s.fakecli.Sub(mes.Subject, fakesub)
		m = s.getCachesub()
		if subs, ok = m[mes.Subject]; ok {
			for _, v := range subs {
				go v(mes)
			}
		}
	}
	return nil
}

func fakesub(m *pubsub.PubsubMsg) {
}

// NewPubSub return new Pubsub instance
func NewPubSub() pubsub.Pubsub {
	return newCli(&server)
}

// GetServer return server instance
func GetServer() Server {
	return &server
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
