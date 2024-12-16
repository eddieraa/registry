package registry

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type services struct {
	_cache map[string][]*Pong
	m      sync.Map
	mu     sync.Mutex
	log    *logrus.Logger
}

func newServices(log *logrus.Logger) *services {
	s := &services{_cache: make(map[string][]*Pong), log: log}
	return s
}

func (s *services) Delete(p *Pong) {
	s.m.Delete(p.Name + p.Address)
}

// DeleteByName delete service bye name with key is serviceName+serviceAddress
func (s *services) DeleteByName(key string) {
	if v, ok := s.m.Load(key); ok {
		s.log.Debug("Delete ok ", key)
		p := v.(*Pong)
		s.m.Delete(key)
		s.rebuildCache(p.Name)
	} else {
		s.log.Debug("Delete not found ", key)
	}
}

func (s *services) IterateAll(f func(key string, p *Pong) bool) {
	s.m.Range(func(k interface{}, v interface{}) bool {
		return f(k.(string), v.(*Pong))
	})
}
func (s *services) Iterate(name string, f func(key string, p *Pong) bool) {
	s.IterateAll(func(k string, p *Pong) bool {
		if p.Name == name {
			return f(k, p)
		}
		return true
	})
}

func (s *services) IterateServiceName(f func(key string) bool) {
	names := make(map[string]bool)
	s.IterateAll(func(key string, p *Pong) bool {
		if ok := names[p.Name]; !ok {
			f(p.Name)
		} else {
			names[p.Name] = true
		}
		return true
	})
}

// LoadOrStore LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (s *services) LoadOrStore(p *Pong) (res *Pong, loaded bool) {
	var v interface{}
	if v, loaded = s.m.LoadOrStore(p.Name+p.Address, p); loaded {
		old := v.(*Pong)
		old.Timestamps = p.Timestamps
		res = old
	} else {
		s.rebuildCache(p.Name)
		res = p
	}
	return res, loaded
}

func (s *services) getCache() map[string][]*Pong {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s._cache
}
func (s *services) setCache(newCache map[string][]*Pong) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s._cache = newCache
}

func (s *services) nbService(name string) (nb int) {
	nb = len(s.getCache()[name])
	return
}
func (s *services) rebuildCache(name string) {
	s.log.Debug("Rebuild cache for ", name)

	ref := s.getCache()
	cache := make(map[string][]*Pong)
	for k, v := range ref {
		cache[k] = v
	}

	if name == "" {
		toDelete := make([]string, 0)
		for k := range cache {
			if _, exist := cache[k]; !exist {
				toDelete = append(toDelete, k)
			}
		}
		if len(toDelete) == 0 {
			for _, k := range toDelete {
				delete(cache, k)
			}
		}
		s.IterateServiceName(func(key string) bool {
			s.rebuildCache(key)
			return true
		})

	} else {
		services := make([]*Pong, 0)
		s.Iterate(name, func(key string, v *Pong) bool {
			services = append(services, v)
			return true
		})
		cache[name] = services
	}
	s.setCache(cache)
}
func (s *services) GetServices(name string) (pongs []*Pong) {
	return s.getCache()[name]
}
