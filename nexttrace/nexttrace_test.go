package nexttrace

import (
	"testing"
	"time"
)

func TestNextTrace3Check(t *testing.T) {
	start := time.Now()
	NextTrace3Check("zh", "ALL", "ipv4")
	duration := time.Since(start)
	t.Logf("执行耗时: %s", duration)
}
