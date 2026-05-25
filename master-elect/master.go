package masterelect

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/eddieraa/registry"
	"github.com/google/uuid"
)

type options struct {
	masterKey string
	idKey     string
	registry  registry.Registry
}

func New(r registry.Registry, serviceName string, opts ...Option) error {
	o := newOptions(opts...)
	o.registry = r

	if o.registry == nil {
		return errors.New("registry is required")
	}

	o.registry.Observe(serviceName)
	o.registry.SetObserverEvent(o.createObserverEvent(serviceName))
	return nil
}

// addIDToService ensures a service has an id in its KV.
// If the id is not defined, a new id is generated with uuid.New() when the service is registered.
// return true if the service was modified, false otherwise
func (o *options) addIDToService(s *registry.Service) bool {
	if s.KV == nil {
		s.KV = make(map[string]string)
	}
	if _, ok := s.KV[o.idKey]; !ok {
		s.KV[o.idKey] = uuid.New().String()
		return true
	}
	return false
}

// create  ObserverEvent func to manage master election when service register/unregister
func (o *options) createObserverEvent(serviceName string) registry.ObserverEvent {
	return func(s registry.Service, event registry.Event) {
		slog.Info("createObserverEvent called with service ", fmt.Sprintf("%+v", s), " and event ", event)
		if s.Name != serviceName {
			return
		}

		if event == registry.EventRegister || event == registry.EventUnregister {
			o.observerEventFunc(serviceName)
			return
		}
	}
}

func (o *options) observerEventFunc(serviceName string) {

	registered, thisService := o.isServiceRegistered(serviceName)
	if !registered {
		return
	}
	if o.addIDToService(&thisService) {
		defer func() {
			o.registry.SetServiceStatus(thisService, registry.Unknown)
		}()
	}
	services, err := o.getServices(serviceName, thisService)
	if err != nil {
		return
	}
	// si thisService n'est pas master, on regarde si thisService est eligible pour devenir master
	// si oui, on update service pour le faire devenir master
	if !o.isMaster(thisService) {
		_, masterFound := o.getMasterService(services)
		if masterFound {
			return
		}
		eligibleMaster, eligibleMasterFound := o.findEligibleMaster(services)
		if !eligibleMasterFound {
			return
		}
		if eligibleMaster.KV[o.idKey] == thisService.KV[o.idKey] {
			thisService.KV[o.masterKey] = "true"
			o.registry.SetServiceStatus(thisService, registry.Unknown)
		}

	}

	// si thisService est master on regarde si dans la liste si il y a d'autres service déjà master
	// si oui on regarde si thisService est toujours eligible, si non on update le service pour qu'il ne le soit plus
	if o.isMaster(thisService) {
		for _, s := range services {
			if s.KV[o.masterKey] == "true" && s.KV[o.idKey] < thisService.KV[o.idKey] {
				thisService.KV[o.masterKey] = "false"
				o.registry.SetServiceStatus(thisService, registry.Unknown)
				return
			}
		}
	}

}

func (o *options) getServices(serviceName string, thisService registry.Service) ([]registry.Service, error) {
	services, err := o.registry.GetServices(serviceName)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, errors.New("no service found")
	}
	// update services with thisService if exist in services list
	// ensure all service have KV with id
	// if not id add id with zzz to ensure this service is not eligible for master election
	for i, s := range services {
		if serviceEqual(s, thisService) {
			services[i] = thisService
			continue
		}
		if s.KV == nil {
			s.KV = make(map[string]string)
		}
		if s.KV[o.idKey] == "" {
			s.KV[o.idKey] = "zzz"
		}
		services[i] = s

	}
	return services, nil
}

func serviceEqual(s1, s2 registry.Service) bool {
	if s1.Name != s2.Name {
		return false
	}
	if s1.Host != s2.Host {
		return false
	}
	if s1.Address != s2.Address {
		return false
	}
	if s1.Network != s2.Network {
		return false
	}
	if s1.URL != s2.URL {
		return false
	}
	compareMap := false
	if compareMap {
		if s1.KV == nil && s2.KV == nil {
			return true
		}
		if (s1.KV == nil && s2.KV != nil) || (s1.KV != nil && s2.KV == nil) {
			return false
		}
		if len(s1.KV) != len(s2.KV) {
			return false
		}
		for k, v := range s1.KV {
			if s2.KV[k] != v {
				return false
			}
		}
	}

	return true
}

func (o *options) isMaster(service registry.Service) bool {
	return service.KV[o.masterKey] == "true"
}

func (o *options) isServiceRegistered(serviceName string) (bool, registry.Service) {
	services := o.registry.GetRegisteredServices()
	for _, s := range services {
		if s.Name == serviceName {
			return true, s
		}
	}
	return false, registry.Service{}
}

func (o *options) getMasterService(services []registry.Service) (registry.Service, bool) {
	for _, s := range services {
		if s.KV[o.masterKey] == "true" {
			return s, true
		}
	}
	return registry.Service{}, false
}

func (o *options) findEligibleMaster(services []registry.Service) (registry.Service, bool) {
	for _, s := range services {
		if o.isEligible(s, services) {
			return s, true
		}
	}
	return registry.Service{}, false
}
func (o *options) isEligible(s registry.Service, services []registry.Service) bool {
	for _, other := range services {
		if other.Name == s.Name && other.KV[o.idKey] < s.KV[o.idKey] {
			return false
		}
	}
	return true
}

func newOptions(opts ...Option) options {
	options := options{
		masterKey: "master",
		idKey:     "id",
		registry:  nil,
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

type Option func(*options)

func WithMasterKey(masterKey string) Option {
	return func(opts *options) {
		opts.masterKey = masterKey
	}
}

func WithIDKey(idKey string) Option {
	return func(opts *options) {
		opts.idKey = idKey
	}
}
