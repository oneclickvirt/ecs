package uts

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func init() {
	setSocketOptions = func(network, address string, c syscall.RawConn, interfaceName string) (err error) {
		switch network {
		case "tcp", "tcp4", "tcp6":
		case "udp", "udp4", "udp6":
		default:
			return
		}
		var innerErr error
		err = c.Control(func(fd uintptr) {
			host, _, _ := net.SplitHostPort(address)
			if ip := net.ParseIP(host); ip != nil && !ip.IsGlobalUnicast() {
				return
			}
			if innerErr = unix.BindToDevice(int(fd), interfaceName); innerErr != nil {
				return
			}
		})
		if innerErr != nil {
			err = innerErr
		}
		return
	}
}
