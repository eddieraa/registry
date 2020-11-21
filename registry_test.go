package registry

import (
	"testing"

	"github.com/sirupsen/logrus"
)

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
