package test

import (
	"testing"

	"github.com/eddieraa/registry"
	"github.com/stretchr/testify/assert"
)

func newFps(name string) func(*registry.PubsubMsg) {
	return func(m *registry.PubsubMsg) {
		println("recv " + name)
	}
}

func Test1(t *testing.T) {
	s := &fakeServer{}
	c1 := &cli{1}
	s.add("test", c1, newFps("c1"))
	c2 := &cli{2}
	s.add("test", c2, newFps("c2"))

	assert.Equal(t, 2, len(s.get("test")))
	s.add("test", c2, newFps("c2"))
	assert.Equal(t, 2, len(s.get("test")))

	s.add("test", &cli{2}, newFps("cx"))
	assert.Equal(t, 2, len(s.get("test")))

	s.add("test", &cli{3}, newFps("cx"))
	assert.Equal(t, 3, len(s.get("test")))
}
