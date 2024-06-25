package cputest

import (
	"testing"
)

func Test(t *testing.T) {
	CpuTest("zh", "sysbench", "1")
}
