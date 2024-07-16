package backtrace

import (
	"net"
	"syscall"

	. "github.com/oneclickvirt/defaultset"
)

func (t *Tracer) listen(network string, laddr *net.IPAddr) (*net.IPConn, error) {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
		conn, err := net.ListenIP(network, laddr)
		if err != nil {
			Logger.Info(err.Error())
			return nil, err
		}
		raw, err := conn.SyscallConn()
		if err != nil {
			Logger.Info(err.Error())
			conn.Close()
			return nil, err
		}
		_ = raw.Control(func(fd uintptr) {
			err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
		})
		if err != nil {
			Logger.Info(err.Error())
			conn.Close()
			return nil, err
		}
		return conn, nil
	} else {
		conn, err := net.ListenIP(network, laddr)
		if err != nil {
			return nil, err
		}
		raw, err := conn.SyscallConn()
		if err != nil {
			conn.Close()
			return nil, err
		}
		_ = raw.Control(func(fd uintptr) {
			err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
		})
		if err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil
	}
}
