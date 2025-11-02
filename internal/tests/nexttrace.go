package tests

import (
	"fmt"
	"strings"

	"github.com/oneclickvirt/nt3/nt"
)

func NextTrace3Check(language, nt3Location, nt3CheckType string) {
	resultChan := make(chan nt.TraceResult, 100)
	go nt.TraceRoute(language, nt3Location, nt3CheckType, resultChan)
	for result := range resultChan {
		if result.Index == -1 {
			for index, res := range result.Output {
				res = strings.TrimSpace(res)
				if res != "" && index == 0 {
					fmt.Println(res)
				}
			}
			continue
		}
		if result.ISPName == "Error" {
			for _, res := range result.Output {
				res = strings.TrimSpace(res)
				if res != "" {
					fmt.Println(res)
				}
			}
			return
		}
		for _, res := range result.Output {
			res = strings.TrimSpace(res)
			if res == "" {
				continue
			}
			if strings.Contains(res, "ICMP") {
				fmt.Print(res)
			} else {
				fmt.Println(res)
			}
		}
	}
}
