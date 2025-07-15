package disktest

import (
	"fmt"
	"testing"
)

func TestDiskIoTest(t *testing.T) {
	_, res := DiskTest("zh", "sysbench", "", false, false)
	fmt.Print(res)
}
