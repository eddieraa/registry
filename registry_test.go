package registry

import (
	"testing"
	"time"

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
	r, err := Connect(Nats(c), Timeout(3000*time.Millisecond))
	if err != nil {
		t.Fatal("Could not open registry session: ", err)
	}

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
	r, err := Connect(Nats(c), Timeout(3000*time.Millisecond), AddFilter(LoadBalanceFilter()))
	if err != nil {
		t.Fatal("Could not open registry session: ", err)
	}
	name := "httptest"
	var s *Service

	s, err = r.GetService(name)
	logrus.Infof("Service %s", s)

	r.Close()
	r, err = Connect(Nats(c), Timeout(3000*time.Millisecond), AddFilter(LoadBalanceFilter()))
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

func newService(addr, host, name string) *Pong {
	return &Pong{Service: Service{Address: addr, Host: host, Name: name}, Timestamps: &Timestamps{Registered: 3, Duration: 5}}
}
func TestChainFilter(t *testing.T) {
	pongs := []*Pong{
		newService("localhost:32", "localhost", "myservice"),
		newService("localhost:34", "localhost", "myservice"),
		newService("localhost:35", "localhost", "myservice"),
		newService("localhost:36", "localhost", "myservice"),
	}
	f := LoadBalanceFilter()

	logrus.Info("\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f))
	logrus.Info("\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f))
	logrus.Info("\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f))
	logrus.Info("\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f), "\n", chainFilters(pongs, f))

}
