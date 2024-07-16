package uts

import (
	"syscall"
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

		if innerErr != nil {
			err = innerErr
		}
		return
	}
}
