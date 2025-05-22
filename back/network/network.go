package network1

import "github.com/oneclickvirt/basics/network"

func NetworkCheck(checkType string, enableSecurityCheck bool, language string) (string, string, error) {
    ipInfo, _, err := network.NetworkCheck(checkType, false, language)
    return ipInfo, "", err
}
