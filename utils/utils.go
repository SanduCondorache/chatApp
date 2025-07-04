package utils

import "net"

func NormalizeAddr(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}

	if host == "::1" {
		host = "127.0.0.1"
	}

	return net.JoinHostPort(host, port)
}
