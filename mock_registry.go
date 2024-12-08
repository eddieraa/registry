package registry

import (
	"errors"
)

// RegistryMock this struct implement interface Registry
//
// implement xxxImpl fields to override your Registry test implementation
type RegistryMock struct {
	RegisterImpl                func(s Service) (FnUnregister, error)
	UnregisterImpl              func(s Service) error
	GetServicesImpl             func(name string) ([]Service, error)
	GetServiceImpl              func(name string, filters ...Filter) (*Service, error)
	ObserveImpl                 func(serviceName string) error
	GetObservedServiceNamesImpl func() []string
	SubscribersImpl             func() []string
	CloseImpl                   func() error
	SetServiceStatusImpl        func(s Service, status Status) error
	GetRegisteredServicesImpl   func() []Service
}

// ErrNotImplemented Not implemented error
var ErrNotImplemented = errors.New("not implemented")

func (r *RegistryMock) Register(s Service) (FnUnregister, error) {
	if r.RegisterImpl != nil {
		return r.RegisterImpl(s)
	}
	return nil, ErrNotImplemented
}
func (r *RegistryMock) Unregister(s Service) error {
	if r.UnregisterImpl != nil {
		return r.UnregisterImpl(s)
	}
	return ErrNotImplemented
}
func (r *RegistryMock) GetServices(name string, opts ...func(*getServicesOptions)) ([]Service, error) {
	if r.GetServicesImpl != nil {
		return r.GetServicesImpl(name)
	}
	return nil, ErrNotImplemented
}
func (r *RegistryMock) GetService(name string, filters ...Filter) (*Service, error) {
	if r.GetServiceImpl != nil {
		return r.GetServiceImpl(name, filters...)
	}
	return nil, ErrNotImplemented
}
func (r *RegistryMock) Observe(serviceName string) error {
	if r.ObserveImpl != nil {
		return r.ObserveImpl(serviceName)
	}
	return ErrNotImplemented
}
func (r *RegistryMock) GetObservedServiceNames() []string {
	if r.GetObservedServiceNamesImpl != nil {
		return r.GetObservedServiceNamesImpl()
	}
	return nil
}
func (r *RegistryMock) Subscribers() []string {
	if r.SubscribersImpl != nil {
		return r.SubscribersImpl()
	}
	return nil
}
func (r *RegistryMock) Close() error {
	if r.CloseImpl != nil {
		return r.CloseImpl()
	}
	return ErrNotImplemented
}
func (r *RegistryMock) SetServiceStatus(s Service, status Status) error {
	if r.SetServiceStatusImpl != nil {
		return r.SetServiceStatusImpl(s, status)
	}
	return ErrNotImplemented
}
func (r *RegistryMock) GetRegisteredServices() []Service {
	if r.GetRegisteredServicesImpl != nil {
		return r.GetRegisteredServicesImpl()
	}
	return nil
}
