package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryTest(t *testing.T) {
	assert.Implements(t, (*Registry)(nil), &RegistryMock{})

	rt := &RegistryMock{}
	rt.GetServiceImpl = func(name string, filters ...Filter) (*Service, error) {
		return &Service{Name: "ok"}, nil
	}
	assert.Implements(t, (*Registry)(nil), rt)
	s, _ := rt.GetService("xx")
	assert.Equal(t, "ok", s.Name)
}

func TestUnimplementedMockRestry(t *testing.T) {
	m := &RegistryMock{}
	assert.ErrorIs(t, ErrNotImplemented, m.Close())
	assert.Nil(t, m.GetObservedServiceNames())
	assert.Nil(t, m.GetRegisteredServices())
	s, err := m.GetService("xxx")
	assert.Nil(t, s)
	assert.ErrorIs(t, ErrNotImplemented, err)
	services, err := m.GetServices("yyyy")
	assert.Nil(t, services)
	assert.ErrorIs(t, ErrNotImplemented, err)
	assert.ErrorIs(t, ErrNotImplemented, m.Observe("tttt"))
	fn, err := m.Register(Service{Name: "myService"})
	assert.Nil(t, fn)
	assert.ErrorIs(t, ErrNotImplemented, err)
	assert.ErrorIs(t, ErrNotImplemented, m.SetServiceStatus(Service{}, Critical))
	assert.Nil(t, m.Subscribers())
	assert.ErrorIs(t, ErrNotImplemented, m.Unregister(Service{}))

}

func TestOverridedMockRegistry(t *testing.T) {
	m := &RegistryMock{
		RegisterImpl:   func(s Service) (FnUnregister, error) { return func() {}, nil },
		UnregisterImpl: func(s Service) error { return nil },
		GetServicesImpl: func(name string) ([]Service, error) {
			return []Service{{Name: "hello"}}, nil
		},
		GetServiceImpl: func(name string, filters ...Filter) (*Service, error) {
			return &Service{Name: "pp"}, nil
		},
		ObserveImpl:                 func(serviceName string) error { return nil },
		GetObservedServiceNamesImpl: func() []string { return []string{"1", "2"} },
		SubscribersImpl:             func() []string { return []string{"A"} },
		CloseImpl:                   func() error { return nil },
		SetServiceStatusImpl:        func(s Service, status Status) error { return nil },
		GetRegisteredServicesImpl:   func() []Service { return nil },
	}

	f, err := m.Register(Service{Name: "toto"})
	assert.NotNil(t, f)
	assert.Nil(t, err)

	assert.Nil(t, m.Unregister(Service{}))

	services, err := m.GetServices("toto")
	assert.Nil(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "hello", services[0].Name)

	s, err := m.GetService("pp")
	assert.Nil(t, err)
	assert.Equal(t, "pp", s.Name)

	assert.Nil(t, m.Observe("yy"))
	assert.NotNil(t, m.GetObservedServiceNames())

	assert.NotNil(t, m.Subscribers())

	assert.Nil(t, m.Close())

	assert.Nil(t, m.SetServiceStatus(Service{Name: "xx"}, Passing))

	assert.Nil(t, m.GetRegisteredServices())

}
