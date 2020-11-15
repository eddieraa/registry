package registry

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type services struct {
	cache map[string][]*Pong
	m     sync.Map
}

func newServices() *services {
	s := &services{cache: make(map[string][]*Pong)}
	return s
}

func (s *services) Delete(p *Pong) {
	s.m.Delete(p.Name + p.Address)
}

//DeleteByName delete service bye name with key is serviceName+serviceAddress
func (s *services) DeleteByName(key string) {
	if v, ok := s.m.Load(key); ok {
		logrus.Debug("Delete ok ", key)
		p := v.(*Pong)
		s.m.Delete(key)
		s.rebuildCache(p.Name)
	} else {
		logrus.Debug("Delete not found ", key)
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

func (s *services) LoadOrStore(p *Pong) (res *Pong) {
	if v, exist := s.m.LoadOrStore(p.Name+p.Address, p); exist {
		old := v.(*Pong)
		old.Timestamps = p.Timestamps
		res = old
	} else {
		s.rebuildCache(p.Name)
		res = p
	}
	return res
}
func (s *services) nbService(name string) (nb int) {
	nb = len(s.cache[name])
	return
}
func (s *services) rebuildCache(name string) {
	logrus.Debug("Rebuild cache for ", name)
	if name == "" {
		toDelete := make([]string, 0)
		for k := range s.cache {
			if _, exist := s.cache[k]; !exist {
				toDelete = append(toDelete, k)
			}
		}
		if len(toDelete) == 0 {
			for _, k := range toDelete {
				delete(s.cache, k)
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
		s.cache[name] = services
	}
}
func (s *services) GetServices(name string) (pongs []*Pong) {
	return s.cache[name]
}
