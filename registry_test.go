package registry

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func Test1(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	c, err := nats.Connect("localhost:4222")
	if err != nil {
		t.Fatal("Could not connect to nats ", err)
	}
	r, err := Connect(c)
	if err != nil {
		t.Fatal("Could not open registry session: ", err)
	}

	ctx, fct := context.WithTimeout(context.TODO(), time.Millisecond*100)
	defer fct()
	services, err := r.GetServices(ctx, "httptest")
	if err != nil {
		t.Error("Could not get services ", err)
		t.Fail()

	}
	assert.NotNil(t, services)
	assert.Equal(t, 2, len(services))
	logrus.Infof("Services %s", services)

	r.GetServices(ctx, "httptest")
	r.GetServices(ctx, "httptest")
	r.GetServices(ctx, "httptest")
	r.GetServices(ctx, "httptest")
	r.GetServices(ctx, "httptest")
	r.GetServices(ctx, "httptest")
	r.GetServices(ctx, "httptest")
	r.GetServices(ctx, "httptest")

}
