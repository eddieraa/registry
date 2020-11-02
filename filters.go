package registry

import "os"

//Filter used for filtering service
type Filter func(services *FilterArg) (accept bool)
type FilterArg struct {
	Nb      int
	Offset  int
	Service Service
	t       Timestamps
}

//LocalhostFilter return true if hostname is equals to service host
func LocalhostFilter() Filter {
	var host string
	host, _ = os.Hostname()

	fn := func(arg *FilterArg) (accept bool) {
		return host == arg.Service.Host
	}
	return fn
}

//LoadBalanceFilter basic loadbalancer
func LoadBalanceFilter() Filter {
	lastInd := -1
	fn := func(arg *FilterArg) (accept bool) {
		lastInd++
		if lastInd >= arg.Nb {
			lastInd = 0
		}
		return arg.Offset == lastInd
	}
	return fn
}
