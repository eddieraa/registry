package registry

import "os"

//Filter used for filtering service
//do not return nil
type Filter func(services []Pong) []Pong

//LocalhostFilter return true if hostname is equals to service host
func LocalhostFilter() Filter {
	var host string
	host, _ = os.Hostname()

	fn := func(services []Pong) []Pong {
		res := []Pong{}
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
	fn := func(services []Pong) []Pong {
		lastInd++
		if len(services) >= lastInd {
			lastInd = 0
		}
		return []Pong{services[lastInd]}
	}
	return fn
}
