package upstreams

import "testing"

func TestUpstreamsCheck(t *testing.T) {
	IPV4 = "148.100.85.25"
	UpstreamsCheck()
}
