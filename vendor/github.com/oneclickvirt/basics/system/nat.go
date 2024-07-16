package system

import (
	"github.com/oneclickvirt/gostun/model"
	"github.com/oneclickvirt/gostun/stuncheck"
)

func getNatType() string {
	model.EnableLoger = false
	addrStrPtrList := []string{
		"stun.voipgate.com:3478",
		"stun.miwifi.com:3478",
		"stunserver.stunprotocol.org:3478",
	}
	checkStatus := true
	for _, addrStr := range addrStrPtrList {
		err1 := stuncheck.MappingTests(addrStr)
		if err1 != nil {
			model.NatMappingBehavior = "inconclusive"
			if model.EnableLoger {
				model.Log.Warn("NAT mapping behavior: inconclusive")
			}
			checkStatus = false
		}
		err2 := stuncheck.FilteringTests(addrStr)
		if err2 != nil {
			model.NatFilteringBehavior = "inconclusive"
			if model.EnableLoger {
				model.Log.Warn("NAT filtering behavior: inconclusive")
			}
			checkStatus = false
		}
		if model.NatMappingBehavior == "inconclusive" || model.NatFilteringBehavior == "inconclusive" {
			checkStatus = false
		} else if model.NatMappingBehavior != "inconclusive" && model.NatFilteringBehavior != "inconclusive" {
			checkStatus = true
		}
		if checkStatus {
			break
		}
	}
	return stuncheck.CheckType()
}
