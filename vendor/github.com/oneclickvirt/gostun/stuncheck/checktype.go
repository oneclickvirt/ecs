package stuncheck

import (
	"fmt"

	"github.com/oneclickvirt/gostun/model"
)

// CheckType
// Summarize the NAT type
func CheckType() string {
	var result string
	if model.NatMappingBehavior != "" && model.NatFilteringBehavior != "" {
		if model.NatMappingBehavior == "inconclusive" || model.NatFilteringBehavior == "inconclusive" {
			result = "Inconclusive"
		} else if model.NatMappingBehavior == "endpoint independent" && model.NatFilteringBehavior == "endpoint independent" {
			result = "Full Cone"
		} else if model.NatMappingBehavior == "endpoint independent" && model.NatFilteringBehavior == "address dependent" {
			result = "Restricted Cone"
		} else if model.NatMappingBehavior == "endpoint independent" && model.NatFilteringBehavior == "address and port dependent" {
			result = "Port Restricted Cone"
		} else if model.NatMappingBehavior == "address and port dependent" && model.NatFilteringBehavior == "address and port dependent" {
			result = "Symmetric"
		} else {
			result = fmt.Sprintf("%v[NatMappingBehavior] %v[NatFilteringBehavior]\n", model.NatMappingBehavior, model.NatFilteringBehavior)
		}
	} else {
		result = "Inconclusive"
	}
	return result
}
