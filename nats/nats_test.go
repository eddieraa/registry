package nats

import (
	"fmt"
	"testing"
	"time"

	"github.com/eddieraa/registry"
	"github.com/eddieraa/registry/pubsub"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const TEST_PORT = 8369

func RunServerOnPort(port int) *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = port
	return RunServerWithOptions(&opts)
}

func RunServerWithOptions(opts *server.Options) *server.Server {
	return natsserver.RunServer(opts)
}

func TestSetDefault(t *testing.T) {
	srv := RunServerOnPort(TEST_PORT)
	defer srv.Shutdown()
	natsURL := fmt.Sprintf("localhost:%d", TEST_PORT)
	c, err := nats.Connect(natsURL)
	if err != nil {
		t.Fatal("Could not connect to nats ", err)
	}
	r, err := SetDefault(c, registry.WithTimeout(3000*time.Millisecond))
	if err != nil {
		t.Fatal("Could not open registry session: ", err)
	}
	r.Close()
}

func TestNewPub(t *testing.T) {
	srv := RunServerOnPort(TEST_PORT)
	defer srv.Shutdown()
	natsURL := fmt.Sprintf("localhost:%d", TEST_PORT)
	c, err := nats.Connect(natsURL)
	if err != nil {
		t.Fatal("Could not connect to nats ", err)
	}
	p := NewPub(c)
	p.Stop()
	goMessage := false
	s, err := p.Sub("test", func(m *pubsub.PubsubMsg) {
		fmt.Println("Got message ", m)
		goMessage = true
	})
	assert.Nil(t, err)
	assert.False(t, goMessage)
	assert.Equal(t, "test", s.Subject())
	defer s.Unsub()
	err = p.Pub("test", []byte("hello"))
	assert.Nil(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.True(t, goMessage)

}

func TestSetLogger(t *testing.T) {

	SetLogLevel(logrus.DebugLevel)
	assert.Equal(t, logrus.DebugLevel, log.Level)
}

func TestWithNatsURL(t *testing.T) {

	natsURL := fmt.Sprintf("localhost:%d", TEST_PORT)
	p := &pb{}
	opts := &registry.Options{
		KVOption: make(map[string]interface{}),
	}
	o := WithNatsUrl(natsURL)
	o(opts)
	p.Configure(opts)
	assert.Equal(t, natsURL, p.natsUrls)

	var srv *server.Server
	defer func() {
		if srv != nil {
			srv.Shutdown()
		}
	}()
	go func() {
		srv = RunServerOnPort(TEST_PORT)
	}()
	p.Sub("xxxxx", func(m *pubsub.PubsubMsg) {})

}
