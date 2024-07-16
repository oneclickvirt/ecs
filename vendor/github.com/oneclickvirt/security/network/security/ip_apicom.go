package security

import (
	"fmt"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// IpApiCom 获取 ip-api.com 的信息
func IpApiCom(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityInfo := &model.SecurityInfo{}
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,mobile,proxy,hosting", ip)
	data, err := utils.FetchJsonFromURL(url, "tcp4", true, "")
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching IpApiCom info: %v", err)
	}
	if isMobile, ok := data["mobile"].(bool); ok {
		securityInfo.IsMobile = utils.BoolToString(isMobile)
	}
	if isProxy, ok := data["proxy"].(bool); ok {
		securityInfo.IsProxy = utils.BoolToString(isProxy)
	}
	if isHosting, ok := data["hosting"].(bool); ok {
		securityInfo.IsDatacenter = utils.BoolToString(isHosting)
	}
	securityInfo.Tag = "5"
	return nil, securityInfo, nil
}
