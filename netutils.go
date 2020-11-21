package registry

import (
	"errors"
	"fmt"
	"net"
)

var ErrNoFreePort = errors.New("No free port available")

//FreePort request a new free port
//return -1 if error
func FreePort() (freeport int, err error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		err = fmt.Errorf("Could not open new port: %v", err)
		return -1, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		err = fmt.Errorf("Error listen addr %s  err: %s", addr, err)
	}
	freeport = l.Addr().(*net.TCPAddr).Port

	defer l.Close()
	return
}

// LocalFreeAddr return local free address (127.0.0.1:34545)
func LocalFreeAddr() (string, error) {
	freeport, err := FreePort()
	if err != nil {
		return "", err
	}
	return fmt.Sprint("127.0.0.1:", freeport), nil
}

//LocalFreeIPv6Addr return local free address ([::1]:34545)
func LocalFreeIPv6Addr() (string, error) {
	freeport, err := FreePort()
	if err != nil {
		return "", err
	}
	return fmt.Sprint("[::1]:", freeport), nil
}

//FindFreePort from range of integer
func FindFreePort(from, to int) (freeport int, err error) {
	var tcpAddr *net.TCPAddr
	for i := from; i <= to; i++ {
		if tcpAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprint("localhost:", i)); err != nil {
			continue
		}
		if l, err := net.ListenTCP("tcp", tcpAddr); err == nil {
			if err = l.Close(); err != nil {
				return -1, err
			}
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return -1, fmt.Errorf("No free port available from range [%d :%d]", from, to)
}

//FindFreeLocalAddress return local free address using range of port
func FindFreeLocalAddress(from, to int) (string, error) {
	freeport, err := FindFreePort(from, to)
	if err != nil {
		return "", err
	}
	return fmt.Sprint("127.0.0.1:", freeport), nil
}
