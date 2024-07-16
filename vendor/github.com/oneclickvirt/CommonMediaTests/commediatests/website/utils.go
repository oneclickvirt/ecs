package website

import (
	"fmt"
	"github.com/oneclickvirt/CommonMediaTests/commediatests/utils"
	"net"
)

func lookupIP(domain string) (string, string, error) {
	var ipv4, ipv6 string
	ns, err := net.LookupHost(domain)
	if err != nil {
		return "", "", fmt.Errorf("net.LookupHost error: %s", err.Error())
	}
	switch {
	case len(ns) != 0:
		for _, n := range ns {
			if utils.ParseIP(n) == 4 {
				ipv4 = n
			}
			if utils.ParseIP(n) == 6 {
				ipv6 = "[" + n + "]"
			}
		}
	}
	return ipv4, ipv6, nil
}

func YoutubeRegionCheck(module string) (string, error) {
	var ipv4, ipv6 string
	dns := "www.youtube.com"
	url := "https://www.youtube.com/red"
	ipv4, ipv6, err := lookupIP(dns)
	if err == nil {
		switch {
		case module == "ipv4":
			return utils.RequestYoutubeIPRegion(url, ipv4), nil
		case module == "ipv6":
			return utils.RequestYoutubeIPRegion(url, ipv6), nil
		default:
			return "", fmt.Errorf("youtube module error: %s", module)
		}
	} else {
		return "", fmt.Errorf(err.Error())
	}
}

func DisneyplusVerifyAuthorized() int {
	ipv4, _, err := lookupIP("global.edge.bamgrid.com")
	if err == nil {
		tokenStatusv4 := utils.RequestDisneyPlusIP("https://global.edge.bamgrid.com/token", ipv4, "auth")
		if tokenStatusv4 == "ok" {
			return 1
		} else if tokenStatusv4 == "400" {
			return -2
		} else {
			return -1
		}
	} else {
		return -1
	}
}

func DisneyplusQueryAreaAvailable(protocol string) string {
	ipv4, ipv6, err := lookupIP("www.disneyplus.com")
	if err == nil {
		switch protocol {
		case "ipv4":
			return utils.CountryCodeToCountryName(utils.RequestDisneyPlusIP("https://www.disneyplus.com", ipv4, "query"))
		case "ipv6":
			return utils.CountryCodeToCountryName(utils.RequestDisneyPlusIP("https://www.disneyplus.com", ipv6, "query"))
		default:
			return ""
		}
	} else {
		return ""
	}
}
