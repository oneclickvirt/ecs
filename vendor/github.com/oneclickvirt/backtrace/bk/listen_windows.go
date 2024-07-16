//go:build windows
// +build windows

package backtrace

import (
	"net"

	. "github.com/oneclickvirt/defaultset"
	"golang.org/x/sys/windows"
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
			err = windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, windows.IP_HDRINCL, 1)
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
			err = windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, windows.IP_HDRINCL, 1)
		})
		if err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil
	}
}
