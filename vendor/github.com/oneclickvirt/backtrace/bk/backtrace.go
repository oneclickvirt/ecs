package backtrace

import (
	"fmt"
	"time"
)

func BackTrace() {
	var (
		s [12]string // 对应 ips 目标地址数量
		c = make(chan Result)
		t = time.After(time.Second * 10)
	)
	for i := range ips {
		go trace(c, i)
	}
loop:
	for range s {
		select {
		case o := <-c:
			s[o.i] = o.s
		case <-t:
			break loop
		}
	}
	for _, r := range s {
		fmt.Println(r)
	}
}
