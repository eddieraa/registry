package registry

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

//Registry Register, Unregister
type Registry interface {
	Register(s Service) (FnUnregister, error)
	GetServices(ctx context.Context, name string) ([]Service, error)
	Observe(serviceName string) error
}

//Pong response to ping
type Pong struct {
	Service
}

//FnUnregister call this func for unregister the service
type FnUnregister func()

//Service service struct
type Service struct {
	Network string
	Address string
	Name    string
	Version string
	Host    string
}

type observe struct {
	callback func(pong Pong)
}

type reg struct {
	c           *nats.Conn
	messageBase string
	m           map[string]map[string]Pong
	observers   map[string]observe
}

func (r reg) buildMessage(message, service string, args ...string) string {
	var b strings.Builder
	b.WriteString(r.messageBase)
	b.WriteString("/")
	b.WriteString(message)
	b.WriteString("/")
	b.WriteString(service)
	if args != nil {
		for _, s := range args {
			b.WriteString("/")
			b.WriteString(s)
		}
	}
	return r.messageBase + "/" + message + "/" + service
}

func (r reg) subToPing(s Service) {
	log.Info("Sub to ping for service ", s.Name, " ", s.Address)
	pong := Pong{
		Service: s,
	}
	fn := func(m *nats.Msg) {
		data, err := json.Marshal(pong)
		if err == nil {
			log.Info("Respond to ping ", m.Reply, " ", s.Name, " ", s.Address)
			m.Respond(data)
		} else {
			log.Errorf("Unable to marchal pong for service %s error is: %s", s.Name, err)
		}

	}
	r.c.Subscribe(r.buildMessage("ping", s.Name), fn)
}

func (r reg) Register(s Service) (f FnUnregister, err error) {
	var data []byte
	f = func() {}
	pong := Pong{Service: s}
	if data, err = json.Marshal(pong); err != nil {
		return
	}
	r.subToPing(s)
	err = r.c.Publish(r.buildMessage("register", s.Name), data)
	if err != nil {
		return
	}
	f = func() {
		r.Unregister(s)
	}
	return

}

func (r reg) Unregister(s Service) (err error) {
	var data []byte
	if data, err = json.Marshal(s); err != nil {
		return
	}
	err = r.c.Publish(r.buildMessage("unregister", s.Name), data)
	return

}

//Connect to NATS
func Connect(c *nats.Conn) (r Registry, err error) {
	r = &reg{c: c, messageBase: "registry", m: make(map[string]map[string]Pong), observers: make(map[string]observe)}
	return
}

func (r reg) getServices(name string) (res []Service) {
	var servicesMap map[string]Pong
	var ok bool
	if servicesMap, ok = r.m[name]; ok && len(servicesMap) > 0 {
		for _, v := range servicesMap {
			res = append(res, v.Service)
		}
		return
	}
	return
}

func (r reg) GetServices(ctx context.Context, name string) ([]Service, error) {
	res := r.getServices(name)
	if res != nil {
		return res, nil
	}
	var observe *observe
	if o, exists := r.observers[name]; !exists {
		observe = &o
		r.Observe(name)
	}
	err := r.c.PublishRequest(r.buildMessage("ping", name), r.buildMessage("register", name), nil)
	if err != nil {
		return nil, err
	}
	ch := make(chan Pong)
	observe.callback = func(pong Pong) {
		ch <- pong
	}

	//Waiting for context done
	select {
	case <-ctx.Done():
		break
	case <-ch:
		break
	}

	res = r.getServices(name)

	return res, nil
}

func (r reg) append(s Pong) {
	if _, ok := r.m[s.Name]; !ok {
		r.m[s.Name] = make(map[string]Pong)
	}
	services := r.m[s.Name]
	services[s.Host+s.Address] = s
}

func (r reg) subregister(msg *nats.Msg) {
	var s Pong
	err := json.Unmarshal(msg.Data, &s)
	if err != nil {
		logrus.Errorf("unmarshal error when sub to register: %s", err)
		return
	}
	r.append(s)
}
func (r reg) subunregister(msg *nats.Msg) {
	var s Service
	err := json.Unmarshal(msg.Data, &s)
	if err != nil {
		logrus.Errorf("unmarshal error when sub to register: %s", err)
		return
	}

	if _, ok := r.m[s.Name]; !ok {
		return
	}
	delete(r.m[s.Name], s.Host+s.Address)

}
func (r reg) Observe(service string) error {
	_, err := r.c.Subscribe(r.buildMessage("register", service), r.subregister)
	if err != nil {
		return err
	}
	_, err = r.c.Subscribe(r.buildMessage("unregister", service), r.subunregister)
	if err != nil {
		return err
	}
	r.observers[service] = observe{}
	return nil
}
