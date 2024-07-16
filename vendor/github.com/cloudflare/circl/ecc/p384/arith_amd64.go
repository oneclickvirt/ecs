//go:build amd64 && !purego
// +build amd64,!purego

package p384

import "golang.org/x/sys/cpu"

var hasBMI2 = cpu.X86.HasBMI2 //nolint
