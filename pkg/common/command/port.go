package command

import (
	"net"
	"strconv"
)

// IsFreePort asks the kernel if the requested port is available or not
func IsFreePort(port int) bool {
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}

	_ = listen.Close()
	return true
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	defer func(listen *net.TCPListener) {
		_ = listen.Close()
	}(listen)
	return listen.Addr().(*net.TCPAddr).Port, nil
}
