package speedtest

import "testing"

func Test(t *testing.T) {
	//NearbySpeedTest("en")
	CustomSpeedTest("https://raw.githubusercontent.com/spiritLHLS/speedtest.net-CN-ID/main/CN_Telecom.csv", 2)
}
