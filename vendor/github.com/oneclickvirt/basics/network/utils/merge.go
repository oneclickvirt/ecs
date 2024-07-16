package utils

import (
	"fmt"

	"github.com/oneclickvirt/basics/model"
)
// chooseString 用于选择非空字符串
func chooseString(src, dst string) string {
	if src != "" {
		return src
	}
	return dst
}

// CompareAndMergeIpInfo 用于比较和合并两个 IpInfo 结构体，非空则不替换
func CompareAndMergeIpInfo(dst, src *model.IpInfo) (res *model.IpInfo, err error) {
	if src == nil {
		return nil, fmt.Errorf("Error merge IpInfo")
	}
	if dst == nil {
		dst = &model.IpInfo{}
	}
	dst.Ip = chooseString(src.Ip, dst.Ip)
	dst.ASN = chooseString(src.ASN, dst.ASN)
	dst.Org = chooseString(src.Org, dst.Org)
	dst.Country = chooseString(src.Country, dst.Country)
	dst.Region = chooseString(src.Region, dst.Region)
	dst.City = chooseString(src.City, dst.City)
	return dst, nil
}

// CompareAndMergeSecurityInfo 用于比较和合并两个 SecurityInfo 结构体，非空则不替换
func CompareAndMergeSecurityInfo(dst, src *model.SecurityInfo) (res *model.SecurityInfo, err error) {
	if src == nil {
		return nil, fmt.Errorf("Error merge SecurityInfo")
	}
	if dst == nil {
		dst = &model.SecurityInfo{}
	}
	dst.IsAbuser = chooseString(src.IsAbuser, dst.IsAbuser)
	dst.IsAttacker = chooseString(src.IsAttacker, dst.IsAttacker)
	dst.IsBogon = chooseString(src.IsBogon, dst.IsBogon)
	dst.IsCloudProvider = chooseString(src.IsCloudProvider, dst.IsCloudProvider)
	dst.IsProxy = chooseString(src.IsProxy, dst.IsProxy)
	dst.IsRelay = chooseString(src.IsRelay, dst.IsRelay)
	dst.IsTor = chooseString(src.IsTor, dst.IsTor)
	dst.IsTorExit = chooseString(src.IsTorExit, dst.IsTorExit)
	dst.IsVpn = chooseString(src.IsVpn, dst.IsVpn)
	dst.IsAnonymous = chooseString(src.IsAnonymous, dst.IsAnonymous)
	dst.IsThreat = chooseString(src.IsThreat, dst.IsThreat)
	dst.Tag = "D"
	return dst, nil
}
