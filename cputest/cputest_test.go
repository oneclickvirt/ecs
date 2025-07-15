package cputest

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	_, res := CpuTest("zh", "sysbench", "1")
	fmt.Print(res)
}
