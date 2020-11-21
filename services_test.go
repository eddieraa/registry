package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func pong(name, address string) *Pong {
	return &Pong{Service: Service{Name: name, Address: address}}
}
func pong2(name, address string, registered int64) *Pong {
	return &Pong{Service: Service{Name: name, Address: address}, Timestamps: &Timestamps{Registered: registered, Duration: 5000}}
}
func TestLoadOrStore(t *testing.T) {
	s := newServices()
	s.LoadOrStore(pong("service1", "localhost:4334"))
	s.LoadOrStore(pong("service1", "localhost:4335"))
	s.LoadOrStore(pong("service1", "localhost:4336"))
	s.LoadOrStore(pong("service1", "localhost:4334"))

	s.LoadOrStore(pong("service2", "localhost:2"))
	s.LoadOrStore(pong2("service2", "localhost:1", 1000))

	assert.Equal(t, 3, s.nbService("service1"))
	assert.Equal(t, 2, s.nbService("service2"))
	var service *Pong
	s.Iterate("service2", func(k string, p *Pong) bool {
		if k == "service2localhost:1" {
			service = p
			return false
		}
		return true
	})
	assert.Equal(t, int64(1000), service.Timestamps.Registered)

	s.LoadOrStore(pong2("service2", "localhost:1", 2000))
	s.Iterate("service2", func(k string, p *Pong) bool {
		if k == "service2localhost:1" {
			service = p
			return false
		}
		return true
	})
	assert.Equal(t, int64(2000), service.Timestamps.Registered)
}
