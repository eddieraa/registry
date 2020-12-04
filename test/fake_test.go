package test

import (
	"testing"

	"github.com/eddieraa/registry"
	"github.com/eddieraa/registry/pubsub"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newFps(name string) func(*pubsub.PubsubMsg) {
	return func(m *pubsub.PubsubMsg) {
		println("recv " + name)
	}
}

func Test1(t *testing.T) {
	s := newFakemap()
	c1 := &cli{1, nil}
	s.add("test", c1, newFps("c1"))
	c2 := &cli{2, nil}
	s.add("test", c2, newFps("c2"))

	assert.Equal(t, 2, len(s.get("test")))
	s.add("test", c2, newFps("c2"))
	assert.Equal(t, 2, len(s.get("test")))

	s.add("test", &cli{2, nil}, newFps("cx"))
	assert.Equal(t, 2, len(s.get("test")))

	s.add("test", &cli{3, nil}, newFps("cx"))
	assert.Equal(t, 3, len(s.get("test")))
}

func TestFake(t *testing.T) {
	nb := 0
	sub := func(pb pubsub.Pubsub) func(m *pubsub.PubsubMsg) {
		return func(m *pubsub.PubsubMsg) {
			nb++
			logrus.Info(pb.(*cli).String()+" rcv "+m.Subject+" data: ", string(m.Data))

		}
	}
	pb := NewPubSub()
	pb.Sub("toto", sub(pb))
	pb.Sub("titi", sub(pb))
	pb3 := NewPubSub()
	pb3.Sub("toto*", sub(pb3))
	pb2 := NewPubSub()
	pb2.Pub("toto", []byte("Hello toto"))
	pb2.Pub("toto", []byte("Hello toto2"))
	pb2.Pub("toto.tutu", []byte("Hello toto"))
	pb2.Pub("titi", []byte("Hello titi"))
	assert.Equal(t, 6, nb)
}

func launchFakeService(t *testing.T) {
	pb := NewPubSub()
	reg, err := registry.NewRegistry(registry.WithPubsub(pb))
	assert.Nil(t, err)
	reg.Register(registry.Service{Name: "test", Address: "localhost:8080"})
	reg.Close()
}

func TestRegistry(t *testing.T) {
	pb := NewPubSub()
	registry.SetDefault(registry.WithPubsub(pb))
	go launchFakeService(t)
	s, err := registry.GetService("test")
	assert.Nil(t, err)
	assert.Equal(t, "localhost:8080", s.Address)
}
