package test

import (
	"sync"

	"github.com/eddieraa/registry"
)

type fakemap struct {
	mu sync.Mutex
	m  map[string]map[cli]func(*registry.PubsubMsg)
}

func newFakemap() *fakemap {
	return &fakemap{m: make(map[string]map[cli]func(*registry.PubsubMsg))}
}
func (s *fakemap) add(topic string, c *cli, f func(m *registry.PubsubMsg)) {
	s.mu.Lock()
	var climap map[cli]func(*registry.PubsubMsg)
	if climap = s.m[topic]; climap == nil {
		climap = make(map[cli]func(*registry.PubsubMsg))
		s.m[topic] = climap
	}
	climap[*c] = f
	s.mu.Unlock()
}
func (s *fakemap) get(topic string) map[cli]func(*registry.PubsubMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m[topic]
}
func (s *fakemap) sizeTopic(topic string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	topics := s.get(topic)
	if topics == nil {
		return 0
	}
	return len(topics)
}

func (s *fakemap) Range(f func(topic string, msgs []func(*registry.PubsubMsg))) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for mess, cliMap := range s.m {
		msgs := make([]func(*registry.PubsubMsg), 0)
		for _, pb := range cliMap {
			msgs = append(msgs, pb)
		}
		f(mess, msgs)

	}
}