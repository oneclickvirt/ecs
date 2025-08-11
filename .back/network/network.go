package network1

import "github.com/oneclickvirt/basics/network"

// 本包在main中不使用
func NetworkCheck(checkType string, enableSecurityCheck bool, language string) (string, string, error) {
	return network.NetworkCheck(checkType, enableSecurityCheck, language)
}
