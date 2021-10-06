package registry

import (
	"errors"
	"fmt"
	"net"
)

//ErrNoFreePort when no free port is available
var ErrNoFreePort = errors.New("No free port available")

//FreePort request a new free port
//return -1 if error
func FreePort() (freeport int, err error) {
	var addr *net.TCPAddr
	addr, err = net.ResolveTCPAddr("tcp", "localhost:0")
	if err == nil {
		var l *net.TCPListener
		l, err = net.ListenTCP("tcp", addr)
		freeport = l.Addr().(*net.TCPAddr).Port
		defer l.Close()
	}
	return
}

// LocalFreeAddr return local free address (127.0.0.1:34545)
func LocalFreeAddr() (freeaddr string, err error) {
	var freeport int
	if freeport, err = FreePort(); err == nil {
		freeaddr = fmt.Sprint("127.0.0.1:", freeport)
	}
	return
}

//LocalFreeIPv6Addr return local free address ([::1]:34545)
func LocalFreeIPv6Addr() (freeaddr string, err error) {
	var freeport int
	if freeport, err = FreePort(); err == nil {
		freeaddr = fmt.Sprint("[::1]:", freeport)
	}
	return
}

//FindFreePort from range of integer
func FindFreePort(from, to int) (freeport int, err error) {
	var tcpAddr *net.TCPAddr
	var fatalerr error
	for i := from; i <= to && fatalerr == nil; i++ {
		if tcpAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprint("localhost:", i)); err == nil {
			if l, err := net.ListenTCP("tcp", tcpAddr); err == nil {
				if fatalerr = l.Close(); fatalerr == nil {
					return l.Addr().(*net.TCPAddr).Port, nil
				}
			}
		}

	}
	return -1, fmt.Errorf("No free port available from range [%d :%d]", from, to)
}

//FindFreeLocalAddress return local free address using range of port (127.0.0.1:23432)
func FindFreeLocalAddress(from, to int) (freeaddr string, err error) {
	var freeport int
	if freeport, err = FindFreePort(from, to); err == nil {
		freeaddr = fmt.Sprint("127.0.0.1:", freeport)
	}
	return
}

//Port extract port from service
func Port(s Service) int {
	var port int
	return port
}
