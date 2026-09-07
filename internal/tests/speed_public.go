//go:build ecs_public

package tests

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/oneclickvirt/speedtest/model"
	"github.com/oneclickvirt/speedtest/sp"
)

func ShowHead(language string) {
	ShowHeadTo(os.Stdout, language)
}

func ShowHeadTo(writer io.Writer, language string) {
	defer func() {
		if recover() != nil {
			fmt.Fprintln(writerOrDiscard(writer), "[WARN] speedtest header unavailable")
		}
	}()
	sp.ShowHeadTo(writerOrDiscard(writer), language)
}

func NearbySP() {
	NearbySPWithNetwork("")
}

func NearbySPWithNetwork(network string) {
	NearbySPWithNetworkTo(os.Stdout, network)
}

func NearbySPWithNetworkTo(writer io.Writer, network string) {
	defer func() {
		if recover() != nil {
			fmt.Fprintln(writerOrDiscard(writer), "[WARN] nearby speedtest unavailable")
		}
	}()
	network = normalizeSpeedNetwork(network)
	if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
		sp.NearbySpeedTestWithNetworkTo(writerOrDiscard(writer), network)
		return
	}
	sp.OfficialNearbySpeedTestWithNetworkTo(writerOrDiscard(writer), network)
}

func writerOrDiscard(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}

// CustomSP keeps public builds on the established public speedtest sources.
func CustomSP(platform, operator string, num int, language string) {
	CustomSPWithNetwork(platform, operator, num, language, "")
}

func CustomSPWithNetwork(platform, operator string, num int, language, network string) {
	CustomSPWithNetworkTo(os.Stdout, platform, operator, num, language, network)
}

func CustomSPWithNetworkTo(writer io.Writer, platform, operator string, num int, language, network string) {
	defer func() {
		if recover() != nil {
			fmt.Fprintln(writerOrDiscard(writer), "[WARN] custom speedtest unavailable")
		}
	}()
	network = normalizeSpeedNetwork(network)

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
		sp.CustomSpeedTestWithNetworkTo(writerOrDiscard(writer), url, parseType, num, language, network)
		return
	}
	sp.OfficialCustomSpeedTestWithNetworkTo(writerOrDiscard(writer), url, parseType, num, language, network)
}

// PrivateSpeedPreloads is a no-op compatibility type for ecs_public builds,
// which intentionally do not link the managed private speed registry.
type PrivateSpeedPreloads struct{}

func StartPrivateSpeedPreloads(context.Context, []string, string) *PrivateSpeedPreloads {
	return &PrivateSpeedPreloads{}
}

func (*PrivateSpeedPreloads) WaitAll(context.Context) error { return nil }

func CustomSPWithNetworkAndPreloads(ctx context.Context, platform, operator string, num int, language, network string, preloads *PrivateSpeedPreloads) {
	CustomSPWithNetworkAndPreloadsTo(os.Stdout, ctx, platform, operator, num, language, network, preloads)
}

func CustomSPWithNetworkAndPreloadsTo(writer io.Writer, _ context.Context, platform, operator string, num int, language, network string, _ *PrivateSpeedPreloads) {
	CustomSPWithNetworkTo(writer, platform, operator, num, language, network)
}
