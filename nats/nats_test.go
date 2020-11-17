package nats

import (
	"testing"
	"time"

	"github.com/eddieraa/registry"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test1(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	c, err := nats.Connect("localhost:4222")
	if err != nil {
		t.Fatal("Could not connect to nats ", err)
	}
	r, err := registry.Connect(Nats(c), registry.Timeout(3000*time.Millisecond))
	if err != nil {
		t.Fatal("Could not open registry session: ", err)
	}
	r.Observe("*")
	services, err := r.GetServices("httptest")
	if err != nil {
		t.Error("Could not get services ", err)
		t.Fail()

	}
	assert.NotNil(t, services)
	assert.Equal(t, 2, len(services))
	logrus.Infof("Services %s", services)
	<-time.Tick(time.Second * 5)
	services, _ = r.GetServices("httptest")
	logrus.Info("nb services ", len(services))
	<-time.Tick(time.Second * 5)
	s, _ := r.GetService("httptest", nil)
	logrus.Infof("Get service %s", s)
	services, _ = r.GetServices("httptest")
	logrus.Info("nb services ", len(services))
	<-time.Tick(time.Second * 5)
	services, _ = r.GetServices("httptest")
	logrus.Info("nb services ", len(services))
	<-time.Tick(time.Second * 5)
	services, _ = r.GetServices("httptest")
	logrus.Info("nb services ", len(services))
	<-time.Tick(time.Second * 5)
	services, _ = r.GetServices("httptest")
	logrus.Info("nb services ", len(services))
	<-time.Tick(time.Second * 5)
	services, _ = r.GetServices("httptest")
	logrus.Info("nb services ", len(services))
	services, _ = r.GetServices("httptest")
	logrus.Info("nb services ", len(services))
	<-time.Tick(time.Second * 5)
	services, _ = r.GetServices("httptest")
	logrus.Info("nb services ", len(services))

	r.Close()
	c.Close()

}

func TestLB(t *testing.T) {
	c, err := nats.Connect("localhost:4222")
	if err != nil {
		t.Fatal("Could not connect to nats ", err)
	}
	r, err := registry.Connect(Nats(c), registry.Timeout(3000*time.Millisecond), registry.AddFilter(registry.LoadBalanceFilter()))
	if err != nil {
		t.Fatal("Could not open registry session: ", err)
	}
	name := "httptest"
	var s *registry.Service

	s, err = r.GetService(name)
	logrus.Infof("Service %s", s)

	r.Close()
	r, err = registry.Connect(Nats(c), registry.Timeout(3000*time.Millisecond), registry.AddFilter(registry.LoadBalanceFilter()))
	if err != nil {
		t.Fatal("Could not open registry session: ", err)
	}

	s, err = r.GetService(name)
	logrus.Infof("Service %s", s.Address)
	s, err = r.GetService(name)
	logrus.Infof("Service %s", s.Address)
	s, err = r.GetService(name)
	logrus.Infof("Service %s", s.Address)
	s, err = r.GetService(name)
	logrus.Infof("Service %s", s.Address)
	s, err = r.GetService(name)
	logrus.Infof("Service %s", s.Address)
	s, err = r.GetService(name)
	logrus.Infof("Service %s", s.Address)
	s, err = r.GetService(name)
	logrus.Infof("Service %s", s.Address)
	s, err = r.GetService(name)
	logrus.Infof("Service %s", s.Address)
	s, err = r.GetService(name)
	logrus.Infof("Service %s", s.Address)
}
