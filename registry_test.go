package registry

import (
	"fmt"
	"strings"
	"testing"
	"time"

	test "github.com/eddieraa/registry/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//create in memory pubsub
var pb = test.NewPubSub()

//var pb pubsub.Pubsub

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})
	/*
		conn, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			logrus.Fatal(err)
		}
		pb = pbnats.NewPub(conn)
	*/
}

func newService(addr, host, name string) *Pong {
	return &Pong{Service: Service{Address: addr, Host: host, Name: name}, Timestamps: &Timestamps{Registered: 3, Duration: 5}}
}

func TestDefaultInstanceNil(t *testing.T) {
	_, err := GetService("test")
	assert.Equal(t, ErrNoDefaultInstance, err)
	_, err = GetServices("test")
	assert.Equal(t, ErrNoDefaultInstance, err)
	err = Observe("test")
	assert.Equal(t, ErrNoDefaultInstance, err)
	_, err = Register(Service{Name: "test"})
	assert.Equal(t, ErrNoDefaultInstance, err)
	err = Unregister(Service{Name: "test"})
	assert.Equal(t, ErrNoDefaultInstance, err)
	err = Close()
	assert.Equal(t, ErrNoDefaultInstance, err)
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

func launchSubscriber(chstop chan interface{}, name string, addr string) {
	reg, _ := NewRegistry(WithPubsub(pb), RegisterInterval(500*time.Millisecond))
	s := Service{Name: name, Address: fmt.Sprint("localhost:", addr)}
	reg.Register(s)
	<-chstop
	reg.Unregister(s)
	reg.Close()
}

func TestRegWithDefaultInstance(t *testing.T) {
	SetDefaultInstance(WithPubsub(pb), RegisterInterval(50*time.Millisecond))
	s, err := GetService("test")
	assert.Nil(t, s)
	assert.NotNil(t, err)

	chstop := make(chan interface{})
	go launchSubscriber(chstop, "tests2", "1")
	go launchSubscriber(chstop, "tests1", "2")
	go launchSubscriber(chstop, "test", "3")

	s, err = GetService("test")
	assert.Nil(t, err)
	assert.NotNil(t, s)
	services, _ := GetServices("test")
	assert.Equal(t, 1, len(services))
	close(chstop)
	Close()

}

func TestWithLB(t *testing.T) {
	r, err := NewRegistry(WithPubsub(pb))
	assert.Nil(t, err)
	err = r.Observe("myservice")
	assert.Nil(t, err)
	chstop := make(chan interface{})
	go launchSubscriber(chstop, "myservice", "11")
	go launchSubscriber(chstop, "myservice", "12")
	go launchSubscriber(chstop, "myservice", "13")
	go launchSubscriber(chstop, "myservice", "14")
	go launchSubscriber(chstop, "myservice", "15")
	go launchSubscriber(chstop, "myservice", "16")
	go launchSubscriber(chstop, "myservice", "17")
	<-time.NewTimer(time.Millisecond * 100).C
	services, err := r.GetServices("myservice")
	assert.Equal(t, 7, len(services))
	assert.Nil(t, err)
	addresses := ""
	lbFilter := LoadBalanceFilter()
	s, err := r.GetService("myservice", lbFilter)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.False(t, strings.Contains(addresses, s.Address))
	addresses += s.Address

	s, _ = r.GetService("myservice", lbFilter)
	assert.False(t, strings.Contains(addresses, s.Address))
	addresses += s.Address

	s, _ = r.GetService("myservice", lbFilter)
	assert.False(t, strings.Contains(addresses, s.Address))
	addresses += s.Address

	s, _ = r.GetService("myservice", lbFilter)
	assert.False(t, strings.Contains(addresses, s.Address))
	addresses += s.Address
	close(chstop)
	r.Close()

}

func TestUnregister(t *testing.T) {
	r, _ := NewRegistry(WithPubsub(pb))
	chstop := make(chan interface{})
	go launchSubscriber(chstop, "testunsub", "1")
	s, _ := r.GetService("testunsub")
	assert.NotNil(t, s)
	chstop <- true
	<-time.NewTimer(time.Millisecond * 10).C
	s, err := r.GetService("testunsub")
	assert.Nil(t, s)
	assert.Equal(t, err, ErrNotFound)
	r.Close()
}

func TestClose(t *testing.T) {
	r, _ := NewRegistry(WithPubsub(pb))
	r.Register(Service{Name: "ppo", Address: "host:345"})
	r.Register(Service{Name: "ppo", Address: "host:346"})
	err := r.Close()
	assert.Nil(t, err)
}

func TestObserveEvent(t *testing.T) {
	chobs := make(chan Event)
	ov := func(s Service, ev Event) {
		chobs <- ev
	}

	r, _ := NewRegistry(WithPubsub(pb), SetObserverEvent(ov))
	r.Observe("testservice")
	chstop := make(chan interface{})
	go launchSubscriber(chstop, "testservice", ":1")
	ev := <-chobs
	assert.Equal(t, EventRegister, ev)
	chstop <- true
	ev = <-chobs
	assert.Equal(t, EventUnregister, ev)
	r.Close()
}

func TestObserveEventWithDefault(t *testing.T) {
	chobs := make(chan Event)
	ov := func(s Service, ev Event) {
		chobs <- ev
	}

	SetDefaultInstance(WithPubsub(pb), SetObserverEvent(ov))
	Observe("testservice2")
	chstop := make(chan interface{})
	go launchSubscriber(chstop, "testservice2", ":1")
	ev := <-chobs
	assert.Equal(t, EventRegister, ev)
	chstop <- true
	ev = <-chobs
	assert.Equal(t, EventUnregister, ev)
	Close()
}

func launchSubscriber2(chstop chan interface{}, name string, addr string) {
	reg, _ := NewRegistry(WithPubsub(pb), RegisterInterval(20*time.Millisecond))
	s := Service{Name: name, Address: fmt.Sprint("localhost:", addr)}
	reg.Register(s)
	<-chstop
	logrus.Info("STTTTOOPPPPPPPP ", name, "    ", addr)

	reg.Unregister(s)
	reg.Close()
}

func TestParalleleSetDefaulInstance(t *testing.T) {
	f := func() {
		SetDefaultInstance(WithPubsub(pb))
		//Close()
	}
	for i := 0; i < 1000; i++ {
		go f()
	}
	Close()
}

func TestCheckDueTime(t *testing.T) {
	SetDefaultInstance(WithPubsub(pb))
	chstop := make(chan interface{})
	go launchSubscriber2(chstop, "checkdutime", "2344")
	go launchSubscriber2(chstop, "checkdutime", "2345")
	go launchSubscriber2(chstop, "checkdutime", "2346")
	go launchSubscriber2(chstop, "checkdutime", "2347")
	go launchSubscriber2(chstop, "checkdutime", "2348")
	go launchSubscriber2(chstop, "checkdutime", "2349")
	go launchSubscriber2(chstop, "checkdutime", "23410")
	go launchSubscriber2(chstop, "checkdutime", "23411")
	go launchSubscriber2(chstop, "checkdutime", "23412")
	go launchSubscriber2(chstop, "checkdutime", "234513")
	s, _ := GetService("checkdutime")
	assert.NotNil(t, s)
	<-time.NewTimer(50 * time.Millisecond).C

	close(chstop)
	Close()

}

func TestPersoFilter(t *testing.T) {
	name := "testpersofilter"
	chstop := make(chan interface{})
	go launchSubscriber(chstop, name, "1")
	go launchSubscriber(chstop, name, "2")
	filter := func(services []*Pong) []*Pong {
		res := []*Pong{}
		for _, s := range services {
			if strings.HasSuffix(s.Address, "2") {
				res = append(res, s)
			}
		}
		return res
	}
	SetDefaultInstance(WithPubsub(pb), AddFilter(filter))
	s, _ := GetService(name)
	assert.NotNil(t, s)
	assert.True(t, strings.HasSuffix(s.Address, "2"))
	close(chstop)
	Close()
}
func TestPongToString(t *testing.T) {
	p := Pong{Service: Service{Name: "test", Address: "localhost:3434", Host: "host"}}
	assert.NotEmpty(t, p.String())
	p = Pong{Service: Service{Name: "test", Address: "localhost:3434", Host: "host"}, Timestamps: &Timestamps{Registered: 21323233433, Duration: 2323234}}
	assert.NotEmpty(t, p.String())

}

func TestErrRegister(t *testing.T) {
	test.GetServer().Pause()
	defer test.GetServer().Resume()

	reg, _ := NewRegistry(WithPubsub(pb))
	_, err := reg.Register(Service{Name: "mys", Address: "x:33"})
	assert.Equal(t, test.ErrPubsubPaused, err)
	reg.Close()

}

func TestSubToPing(t *testing.T) {
	service := "subtoping"
	test.GetServer().Resume()
	SetDefaultInstance(WithPubsub(pb))
	_, err := Register(Service{Name: service, Address: "x:123"})
	assert.Nil(t, err)

	reg, _ := NewRegistry(WithPubsub(pb))

	s, _ := reg.GetService(service)
	assert.NotNil(t, s)

	reg.Close()
	Close()
}
func TestDueTime(t *testing.T) {
	service := "service-checkduetime"
	test.GetServer().Resume()
	SetDefaultInstance(WithPubsub(pb), RegisterInterval(20*time.Millisecond))
	ch := make(chan interface{})
	go launchSubscriber(ch, service, "h:43")
	s, _ := GetService(service)
	assert.NotNil(t, s)
	test.GetServer().Pause()
	<-time.NewTimer(1000 * time.Millisecond).C
	s, err := GetService(service)
	assert.NotNil(t, err)
	assert.Nil(t, s)

	close(ch)
	Close()
}

func TestMainTopic(t *testing.T) {
	r, _ := NewRegistry(WithPubsub(pb), MainTopic("maintopic"))
	assert.Equal(t, "maintopic.toto.titi", r.(*reg).buildMessage("toto", "titi"))
	r.Close()

}
