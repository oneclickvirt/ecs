package memorytest

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	_, res := MemoryTest("zh", "sysbench")
	fmt.Print(res)
}
