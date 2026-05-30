package masterelect

import (
	"testing"

	"github.com/eddieraa/registry"
	"github.com/stretchr/testify/assert"
)

func TestObserverEventFunc(t *testing.T) {
	services := map[string]*registry.Service{}
	r := &registry.RegistryMock{
		SetServiceStatusImpl: func(s registry.Service, status registry.Status) error {
			services[s.KV["id"]] = &s
			return nil
		},
		GetServicesImpl: func(name string) ([]registry.Service, error) {
			return []registry.Service{*services["1"], *services["2"]}, nil
		},
	}

	o := &options{
		masterKey: "master",
		idKey:     "id",
		registry:  r,
	}
	//
	// test with no registered service
	//
	services = map[string]*registry.Service{
		"1": {Name: "myService", KV: map[string]string{"id": "1"}},
		"2": {Name: "myService", KV: map[string]string{"id": "2"}},
	}
	r.GetRegisteredServicesImpl = func() []registry.Service {
		return []registry.Service{}
	}
	o.observerEventFunc("myService")
	assert.False(t, o.isMaster(*services["1"]))
	assert.False(t, o.isMaster(*services["2"]))

	//
	// service 1 register, service 1 should be master
	//
	services = map[string]*registry.Service{
		"1": {Name: "myService", KV: map[string]string{"id": "1"}},
		"2": {Name: "myService", KV: map[string]string{"id": "2"}},
	}
	r.GetRegisteredServicesImpl = func() []registry.Service {
		return []registry.Service{*services["1"]}
	}
	o.observerEventFunc("myService")
	assert.True(t, o.isMaster(*services["1"]))
	assert.False(t, o.isMaster(*services["2"]))

	//
	// service 2 registered, no change because service 1 has lower id
	//
	services = map[string]*registry.Service{
		"1": {Name: "myService", KV: map[string]string{"id": "1"}},
		"2": {Name: "myService", KV: map[string]string{"id": "2"}},
	}
	r.GetRegisteredServicesImpl = func() []registry.Service {
		return []registry.Service{*services["2"]}
	}
	o.observerEventFunc("myService")
	assert.False(t, o.isMaster(*services["1"]))
	assert.False(t, o.isMaster(*services["2"]))

}

func TestIsEligible(t *testing.T) {
	//
	// services with no KV, ensure no panic
	//
	services := map[string]*registry.Service{}
	r := &registry.RegistryMock{
		SetServiceStatusImpl: func(s registry.Service, status registry.Status) error {
			services[s.KV["id"]] = &s
			return nil
		},
		GetServicesImpl: func(name string) ([]registry.Service, error) {
			return []registry.Service{*services["1"], *services["2"]}, nil
		},
	}

	o := &options{
		masterKey: "master",
		idKey:     "id",
		registry:  r,
	}

	services = map[string]*registry.Service{
		"1": {Name: "myService", Host: "h1", Address: "a1"},
		"2": {Name: "myService", Host: "h2", Address: "a2"},
	}
	r.GetRegisteredServicesImpl = func() []registry.Service {
		return []registry.Service{*services["2"]}
	}
	o.observerEventFunc("myService")
	assert.False(t, o.isMaster(*services["1"]))
	assert.True(t, o.isMaster(*services["2"]))
}
