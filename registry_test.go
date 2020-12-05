package registry

import (
	"fmt"
	"net"
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
	myaddr := addr
	host := ""
	if strings.Contains(addr, ":") {
		host = strings.Split(addr, ":")[0]
	} else {
		myaddr = fmt.Sprint("localhost:", addr)
	}
	s := Service{Name: name, Address: myaddr, Host: host}
	fn, _ := reg.Register(s)
	<-chstop
	if fn != nil {
		fn()
	}
	reg.Close()
}

func TestRegWithDefaultInstance(t *testing.T) {
	SetDefault(WithPubsub(pb), RegisterInterval(50*time.Millisecond))
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
	reset()
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

	SetDefault(WithPubsub(pb), SetObserverEvent(ov))
	Observe("testservice2")
	chstop := make(chan interface{})
	go launchSubscriber(chstop, "testservice2", "1")
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
		SetDefault(WithPubsub(pb))
		//Close()
	}
	for i := 0; i < 1000; i++ {
		go f()
	}
	Close()
}

func TestCheckDueTime(t *testing.T) {
	SetDefault(WithPubsub(pb))
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

func TestFilters(t *testing.T) {
	//test chainFilter
	name := "testfilter"
	pongs := []*Pong{
		newService("10.11.11.10:3434", "10.11.11.10", name),
		newService("10.11.11.10:3435", "10.11.11.10", name),
		newService("10.11.11.10:3436", "10.11.11.10", name),
		newService("10.11.11.11:3434", "10.11.11.10", name),
		newService("10.11.11.12:3434", "10.11.11.10", name),
		newService("10.11.11.13:3434", "10.11.11.10", name),
	}
	ser := chainFilters(pongs, LocalhostFilter())
	assert.NotNil(t, ser)
	assert.Empty(t, ser)
	reset()

	chstop := make(chan interface{})
	go launchSubscriber(chstop, name, "10.11.1.11:5454")
	go launchSubscriber(chstop, name, "10.11.1.12:5454")
	SetDefault(WithPubsub(pb), AddFilter(LocalhostFilter()))
	services, err := GetService(name)
	assert.Nil(t, services)
	assert.Equal(t, ErrNotFound, err)
	close(chstop)
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
	SetDefault(WithPubsub(pb), AddFilter(filter))
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
	SetDefault(WithPubsub(pb))
	_, err := Register(Service{Name: service, Address: "x:123"})
	assert.Nil(t, err)

	reg, _ := NewRegistry(WithPubsub(pb))

	s, _ := reg.GetService(service)
	assert.NotNil(t, s)

	reg.Close()
	Close()
}
func TestDueTime(t *testing.T) {
	reset()
	service := "service-checkduetime"
	test.GetServer().Resume()
	SetDefault(WithPubsub(pb), RegisterInterval(20*time.Millisecond), Timeout(60*time.Millisecond))
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
	reset()
}

func TestMainTopic(t *testing.T) {
	r, _ := NewRegistry(WithPubsub(pb), MainTopic("maintopic"))
	assert.Equal(t, "maintopic.toto.titi", r.(*reg).buildMessage("toto", "titi"))
	r.Close()

}

func TestGetSubscribers(t *testing.T) {
	reset()
	_, err := GetDefault()
	assert.Equal(t, err, ErrNoDefaultInstance)
	SetDefault(WithPubsub(pb))
	r, err := GetDefault()
	assert.Nil(t, err)
	services := r.Subscribers()
	assert.Empty(t, services)
	ch := make(chan interface{})
	go launchSubscriber(ch, "testregistered.s1", "h:43")
	go launchSubscriber(ch, "testregistered.s1", "h:44")
	go launchSubscriber(ch, "testregistered.s2", "h:43")
	go launchSubscriber(ch, "testregistered.s3", "h:43")
	go launchSubscriber(ch, "testregistered.s4", "h:43")
	r.Observe("testregistered.s1")
	services = r.Subscribers()
	assert.Equal(t, 1, len(services))
	r.Observe("testregistered.s2")
	services = r.Subscribers()
	assert.Equal(t, 2, len(services))
	r.Close()
	SetDefault(WithPubsub(pb))
	r, _ = GetDefault()
	services = r.Subscribers()
	assert.Equal(t, 0, len(services))

	close(ch)
}

func TestGetSubscribers2(t *testing.T) {
	reset()
	_, err := GetDefault()
	assert.Equal(t, err, ErrNoDefaultInstance)
	SetDefault(WithPubsub(pb))
	r, err := GetDefault()
	assert.Nil(t, err)
	services := r.Subscribers()
	assert.Empty(t, services)
	ch := make(chan interface{})
	go launchSubscriber(ch, "testregistered2.s1", "h:43")
	go launchSubscriber(ch, "testregistered2.s1", "h:44")
	go launchSubscriber(ch, "testregistered2.s2", "h:43")
	go launchSubscriber(ch, "testregistered2.s3", "h:43")
	go launchSubscriber(ch, "testregistered2.s4", "h:43")
	r.Observe("testregistered2.*")
	<-time.NewTimer(1000 * time.Millisecond).C
	services = r.Subscribers()
	assert.Equal(t, 4, len(services))
	r.Observe("testregistered2.s3")
	<-time.NewTimer(100 * time.Millisecond).C
	assert.Equal(t, 4, len(services))
	close(ch)

}

func TestFindFreePort(t *testing.T) {
	min := 10000
	max := 10100
	p, err := FindFreePort(min, max)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, p, min)
	assert.LessOrEqual(t, p, max)

	addr, err := LocalFreeAddr()
	assert.Nil(t, err)
	assert.NotEmpty(t, addr)

	addr, err = FindFreeLocalAddress(min, max)
	assert.Nil(t, err)
	assert.NotEmpty(t, addr)
	var n *net.TCPAddr
	if n, err = net.ResolveTCPAddr("tcp", addr); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", n); err == nil {
			addr, err = FindFreeLocalAddress(min, min)
			l.Close()
		}
	}

	assert.NotNil(t, err)
	assert.Empty(t, addr)
}

func reset() {
	Close()
	test.GetServer().Resume()
	debug := pb.(test.Debug)

	debug.CallbackPub(nil)
	debug.CallbackSub(nil)
}
func TestOptsTimeout(t *testing.T) {
	reset()
	mypb := test.NewPubSub()
	r, _ := NewRegistry(WithPubsub(mypb), Timeout(50*time.Millisecond))
	now := time.Now()
	s, err := r.GetService("testtimeout")
	assert.NotNil(t, err)
	assert.Empty(t, s)
	d := time.Now().Sub(now) - 50*time.Millisecond

	assert.Greater(t, int64(d), int64(0))
	Close()
}

func TestAddObserveFilter(t *testing.T) {
	reset()
	of := func(p *Pong) (res bool) {
		if strings.HasPrefix(p.Address, "localhost:") {
			res = true
		}
		logrus.Debug("filter ", p.Address, " res ", res)
		return
	}
	SetDefault(WithPubsub(pb), AddObserveFilter(of))
	ch := make(chan interface{})
	go launchSubscriber(ch, "of.s1", "localhost:43")
	go launchSubscriber(ch, "of.s1", "localhost:44")
	go launchSubscriber(ch, "of.s1", "10.10.1.11:43")
	go launchSubscriber(ch, "of.s3", "10.10.1.11:43")
	Observe("of.s1")
	<-time.NewTimer(500 * time.Millisecond).C
	services, _ := GetServices("of.s1")
	assert.Equal(t, 2, len(services))
	close(ch)
}

func TestLocalFreeIPv6Addr(t *testing.T) {
	add, err := LocalFreeIPv6Addr()
	assert.Nil(t, err)
	assert.NotNil(t, add)
}

func TestConcurrentAccessToRegisterdServices(t *testing.T) {
	reset()
	SetDefault(WithPubsub(pb), RegisterInterval(time.Millisecond*1))
	count := 0
	ch := make(chan bool)
	registerFn := func() {
		count++
		serviceName := fmt.Sprint("sreg", count)
		Register(Service{Name: serviceName, Address: "localhost:2134"})
		<-ch
	}
	registry, _ := GetDefault()
	r := registry.(*reg)
	for n := 0; n < 10; n++ {
		go r.registerServiceInContinue()

	}

	count2 := 0
	unRegisterFn := func() {
		count2++
		serviceName := fmt.Sprint("sreg", count2)
		Unregister(Service{Name: serviceName, Address: "localhost:2134"})
		<-ch
	}
	for n := 0; n < 50; n++ {
		go registerFn()
		go unRegisterFn()

	}
	<-time.NewTimer(100 * time.Millisecond).C
	close(ch)
	Close()
}

func TestMarshal(t *testing.T) {
	pb.(test.Debug).CallbackPub(func(s string, b []byte) ([]byte, error) {
		return []byte("titi toto"), nil
	})
	SetDefault(WithPubsub(pb))
	ch := make(chan interface{})
	go launchSubscriber(ch, "test", "localhost:43")

	s, err := GetService("test")
	assert.Nil(t, s)
	assert.NotNil(t, err)

	close(ch)
	reset()
}

func TestLocalhostOFilter(t *testing.T) {
	reset()
	SetDefault(WithPubsub(pb), AddObserveFilter(LocalhostOFilter()))
	ch := make(chan interface{})
	go launchSubscriber(ch, "test", "43")

	s, err := GetService("test")
	close(ch)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	<-time.NewTimer(50 * time.Millisecond).C
	s, err = GetService("test")
	assert.NotNil(t, err)
	assert.Nil(t, s)
	ch = make(chan interface{})
	go launchSubscriber(ch, "test", "10.1.10.4:43")
	s, err = GetService("test")
	assert.NotNil(t, err)
	assert.Nil(t, s)
}
