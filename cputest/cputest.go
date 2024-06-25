package cputest

import (
	"fmt"
	"github.com/oneclickvirt/cputest/cpu"
)

func cputest() {
	//res := cpu.SysBenchTest("zh", "1")
	res := cpu.WinsatTest("zh", "1")
	fmt.Println(res)
}
