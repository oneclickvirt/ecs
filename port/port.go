package port

import (
	"fmt"
	"github.com/oneclickvirt/portchecker/email"
)

// 常用端口阻断检测 TCP/UDP/ICMP 协议
func portcheck() {
	res := email.EmailCheck()
	fmt.Println(res)
}
