//go:build !ecs_public
package tests

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/oneclickvirt/speedtest/model"
	"github.com/oneclickvirt/speedtest/sp"
)

func ShowHead(language string) {
	defer func() {
		if recover() != nil {
			fmt.Fprintln(os.Stderr, "[WARN] speedtest header unavailable")
		}
	}()
	sp.ShowHead(language)
}

func NearbySP() {
	defer func() {
		if recover() != nil {
			fmt.Fprintln(os.Stderr, "[WARN] nearby speedtest unavailable")
		}
	}()
	if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
		sp.NearbySpeedTest()
		return
	}
	sp.OfficialNearbySpeedTest()
}

// CustomSP keeps public builds on the established public speedtest sources.
func CustomSP(platform, operator string, num int, language string) {
	defer func() {
		if recover() != nil {
			fmt.Fprintln(os.Stderr, "[WARN] custom speedtest unavailable")
		}
	}()

	var url, parseType string
	switch strings.ToLower(platform) {
	case "cn":
		switch strings.ToLower(operator) {
		case "cmcc":
			url = model.CnCMCC
		case "cu":
			url = model.CnCU
		case "ct":
			url = model.CnCT
		case "hk":
			url = model.CnHK
		case "tw":
			url = model.CnTW
		case "jp":
			url = model.CnJP
		case "sg":
			url = model.CnSG
		}
		parseType = "url"
	case "net":
		switch strings.ToLower(operator) {
		case "cmcc":
			url = model.NetCMCC
		case "cu":
			url = model.NetCU
		case "ct":
			url = model.NetCT
		case "hk":
			url = model.NetHK
		case "tw":
			url = model.NetTW
		case "jp":
			url = model.NetJP
		case "sg":
			url = model.NetSG
		case "global", "other":
			url = model.NetGlobal
		}
		parseType = "id"
	}
	if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
		sp.CustomSpeedTest(url, parseType, num, language)
		return
	}
	sp.OfficialCustomSpeedTest(url, parseType, num, language)
}
