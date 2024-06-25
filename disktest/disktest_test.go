package disktest

import "testing"

func TestDiskIoTest(t *testing.T) {
	DiskTest("zh", "sysbench", "", false)
}
