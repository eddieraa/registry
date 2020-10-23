package registry

import "os"

//Filter used for filtering service
type Filter func(s Service) bool

//LocalhostFilter return true if hostname is equals to service host
func LocalhostFilter(hostname string) Filter {
	var host string
	if hostname == "" {
		host, _ = os.Hostname()
	}

	fn := func(s Service) bool {
		if hostname == s.Host {
			return true
		}
		if host == s.Host {
			return true
		}
		return false
	}
	return fn
}
