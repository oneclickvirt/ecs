package memorytest

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	_, res := MemoryTest("zh", "stream")
	fmt.Print(res)
}
