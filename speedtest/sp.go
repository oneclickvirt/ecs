package speedtest

import (
	"github.com/oneclickvirt/speedtest/model"
	"github.com/oneclickvirt/speedtest/sp"
	"runtime"
	"strings"
)

func ShowHead(language string) {
	sp.ShowHead(language)
}

func NearbySP() {
	if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
		sp.NearbySpeedTest()
	} else {
		sp.OfficialNearbySpeedTest()
	}
}

func CustomSP(platform, operator string, num int, language string) {
	var url, parseType string
	if strings.ToLower(platform) == "cn" {
		if strings.ToLower(operator) == "cmcc" {
			url = model.CnCMCC
		} else if strings.ToLower(operator) == "cu" {
			url = model.CnCU
		} else if strings.ToLower(operator) == "ct" {
			url = model.CnCT
		} else if strings.ToLower(operator) == "hk" {
			url = model.CnHK
		} else if strings.ToLower(operator) == "tw" {
			url = model.CnTW
		} else if strings.ToLower(operator) == "jp" {
			url = model.CnJP
		} else if strings.ToLower(operator) == "sg" {
			url = model.CnSG
		}
		parseType = "url"
	} else if strings.ToLower(platform) == "net" {
		if strings.ToLower(operator) == "cmcc" {
			url = model.NetCMCC
		} else if strings.ToLower(operator) == "cu" {
			url = model.NetCU
		} else if strings.ToLower(operator) == "ct" {
			url = model.NetCT
		} else if strings.ToLower(operator) == "hk" {
			url = model.NetHK
		} else if strings.ToLower(operator) == "tw" {
			url = model.NetTW
		} else if strings.ToLower(operator) == "jp" {
			url = model.NetJP
		} else if strings.ToLower(operator) == "sg" {
			url = model.NetSG
		} else if strings.ToLower(operator) == "global" {
			url = model.NetGlobal
		}
		parseType = "id"
	}
	if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
		sp.CustomSpeedTest(url, parseType, num, language)
	} else {
		sp.OfficialCustomSpeedTest(url, parseType, num, language)
	}
}
