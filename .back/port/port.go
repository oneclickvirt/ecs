package port

import (
	"fmt"
	"github.com/oneclickvirt/portchecker/email"
)

// 常用端口阻断检测 TCP/UDP/ICMP 协议
// 本包不在main中使用
func EmailCheck() {
	res := email.EmailCheck()
	fmt.Println(res)
}
