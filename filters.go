package registry

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Filter used for filtering service
// do not return nil
// Ex Load Balancing
type Filter func(services []*Pong) []*Pong

// ObserveFilter used to accept (or not) new registered service.
//
// Ex: you want only service on same host
type ObserveFilter func(s *Pong) bool

func hostname() string {
	host, err := os.Hostname()
	if err != nil {
		logrus.Error("Unable to get local hostname: ", err.Error())
	}
	return host
}

// LocalhostFilter return true if hostname is equals to service host
func LocalhostFilter() Filter {
	host := hostname()

	fn := func(services []*Pong) []*Pong {
		res := []*Pong{}
		for _, s := range services {
			if s.Host == host {
				res = append(res, s)
			}

		}
		return res
	}
	return fn
}

// LoadBalanceFilter basic loadbalancer
func LoadBalanceFilter() Filter {
	lastInd := -1
	emptyServices := []*Pong{}
	fn := func(services []*Pong) []*Pong {
		lastInd++
		size := len(services)
		if size == 0 {
			return emptyServices
		}
		if lastInd >= size {
			lastInd = 0
		}
		return []*Pong{services[lastInd]}
	}
	return fn
}

func PassingFilter() Filter {
	fn := func(services []*Pong) []*Pong {
		res := []*Pong{}
		for _, p := range services {
			if p.Status == Passing {
				res = append(res, p)
			}
		}
		return res
	}
	return fn
}

// LocalhostOFilter accept only service on the same machine
func LocalhostOFilter() ObserveFilter {
	name := hostname()
	return func(p *Pong) bool {
		return name == p.Host
	}
}
