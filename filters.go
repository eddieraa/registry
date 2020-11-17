package registry

import "os"

//Filter used for filtering service
//do not return nil
//Ex Load Balancing
type Filter func(services []*Pong) []*Pong

//ObserveFilter used to accept (or not) new registered service.
//
//Ex: you want only service on same host
type ObserveFilter func(s *Pong) bool

//LocalhostFilter return true if hostname is equals to service host
func LocalhostFilter() Filter {
	var host string
	host, _ = os.Hostname()

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

//LoadBalanceFilter basic loadbalancer
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
