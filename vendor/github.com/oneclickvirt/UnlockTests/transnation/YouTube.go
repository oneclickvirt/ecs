package transnation

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// Youtube
// www.youtube.com 双栈 且 get 请求
func Youtube(c *http.Client) model.Result {
	name := "YouTube Region"
	hostname := "www.youtube.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.youtube.com/premium"
	headers := map[string]string{
		"User-Agent": model.UA_Browser,
		"Cookie":     "YSC=BiCUU3-5Gdk; CONSENT=YES+cb.20220301-11-p0.en+FX+700; GPS=1; VISITOR_INFO1_LIVE=4VwPMkB7W5A; SOCS=CAISOAgDEitib3FfaWRlbnRpdHlmcm9udGVuZHVpc2VydmVyXzIwMjQwNTIxLjA3X3AxGgV6aC1DTiACGgYIgNTEsgY; PREF=f7=4000&tz=Asia.Shanghai&f4=4000000; _gcl_au=1.1.1809531354.1646633279",
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body1 := string(b)
	if strings.Contains(body1, "www.google.cn") {
		return model.Result{Name: name, Status: model.StatusNo, Region: "cn"}
	}
	if strings.Contains(body1, "Premium is not available in your country") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if EndLocation := strings.Index(body1, `"countryCode":`); EndLocation != -1 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{
			Name: name, Status: model.StatusYes,
			Region:     strings.ToLower(body1[EndLocation+15 : EndLocation+17]),
			UnlockType: unlockType,
		}
	}
	if strings.Contains(body1, "premiumPurchaseButton") ||
		strings.Contains(body1, "manageSubscriptionButton") ||
		strings.Contains(body1, "/月") || strings.Contains(body1, "/month") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType, Region: "us"}
	}
	return model.Result{Name: name, Status: model.StatusNo}
}

// YoutubeCDN
// redirector.googlevideo.com 双栈 且 get 请求
func YoutubeCDN(c *http.Client) model.Result {
	name := "YouTube CDN"
	hostname := "googlevideo.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://redirector.googlevideo.com/report_mapping"
	client := utils.Req(c)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body := string(b)
	i := strings.Index(body, "=> ")
	if i == -1 {
		return model.Result{Name: name, Status: model.StatusUnexpected}
	}
	body = body[i+3:]
	i = strings.Index(body, " ")
	if i == -1 {
		return model.Result{Name: name, Status: model.StatusUnexpected}
	}
	body = body[:i]
	i = strings.Index(body, "-")

	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	if i == -1 {
		i = strings.Index(body, ".")
		return model.Result{
			Name: name, Status: model.StatusYes, Info: "Youtube Video Server",
			Region:     findAirCode(body[i+1:]),
			UnlockType: unlockType,
		}
	} else {
		isp := body[:i]
		return model.Result{
			Name: name, Status: model.StatusYes, Info: "Google Global CacheCDN - ISP Cooperation",
			Region:     isp + " - " + findAirCode(body[i+1:]),
			UnlockType: unlockType,
		}
	}
}

func findAirCode(code string) string {
	airPortCode := []string{
		"KIX", "NRT", "GMP", "YOW", "YMQ/YUL", "YVR", "YYC", "YEG", "YTO/YYZ", "WAS/IAD", "ABE", "ABQ", "ATL", "AUS", "AZO", "BDL", "BHM", "BNA", "BOI", "BOS", "BRO", "BTR", "BTL", "BUF", "BWI", "CAE", "CAK", "CHA", "CHI/ORD", "CHS", "CID", "CLE", "CLT", "CMH", "CRP", "CVG", "DAY", "DEN", "DFW", "DSM", "DTW", "ELP", "ERI", "EWR", "EVV", "FLL", "FNT", "FWA", "GRR", "GEG", "GSO", "GSP", "GRB", "HAR", "HOU/IAH", "HSV", "HNL", "ICT", "ILM", "IND", "JAN", "JAX", "LAS", "LAX", "LEX", "LIT", "LNK", "LRD", "MCI", "MCO", "MEM", "MFE", "MIA", "MKC", "MKE", "MSN", "MSP", "MSY", "MOB", "NYC/JFK", "OKC", "OMA", "ORF", "ORL", "PBI", "PDX", "PHL/PHA", "PHX", "PIA", "PIT", "PNS", "PVD", "RDU", "RIC", "RNO", "ROC", "SAN", "SAT", "SAV", "SBN", "SDF", "SEA", "BFI", "SFO", "SGF", "SHV", "SLC", "SMF", "STL", "TUL", "SYR", "TOL", "TPA", "TUL", "TUS", "TYS", "MEX", "GDL/MEX", "GUA", "TGU", "SAL", "MGA", "SJO", "PTY", "NAS", "HAV", "SCU", "KIN", "PAP", "SDQ", "SJU", "ROX", "GND", "BGI", "POS", "BOG", "CCS", "GEO", "PBM", "CAY", "BSB", "CWB", "POA", "MAO", "RIO", "SAO", "UIO", "GYE", "LIM", "SRE", "ASU", "MVD", "BUE", "ANF", "SCL", "PTP", "LON/LHR", "ABZ", "BHX", "BOH", "BRS", "CWL", "EDI", "EXT", "GLA", "LPL", "MAN", "NWI", "PLH", "SOU", "BRS", "CDQ", "CVT", "LBA", "PME", "NCL", "HUY", "PIK", "EMA", "BFS", "DUB", "ORK", "SNN", "BRU", "ANR", "OST", "LUX", "AMS", "RTM", "EIN", "ENS", "CPH", "ALL", "AAR", "BLL", "BER/TXL", "MUC", "BRE", "HAJ", "DUS", "FRA", "LEJ", "DUI", "STR", "HAM", "ERF", "FMO", "NUE", "DRS", "SCN", "CGN", "DTM", "BFE", "ZTZ", "ESS", "BON", "RUN", "PAR/CDG", "MRS", "LYS", "BOD", "LIL", "TLS", "NTE", "MLH", "MPL", "GNB", "URO", "NCE", "SXB", "XVE", "PPT", "XMM/GRZ", "BRN", "GVA", "ZRH", "BSL", "ALV", "MAD", "ALC", "BCN", "VLC", "SVQ", "AGP", "VLL", "LIS", "OPO", "ROM", "AHO", "AOI", "BDS", "BLQ", "BRI", "GOA", "MIL/MXP", "SWK", "NAP", "VCE", "FLR", "TRN", "TRS", "CTA", "TAR", "PSA", "QME", "VRN", "ATH", "SKG", "VIE", "LNZ", "GRZ", "SZG", "INN", "PRG", "HEL", "STO/ARN", "AGH", "GOT", "MMA/MMX", "NRK", "OSL", "TIA", "SKP", "SOF", "BEG", "BUH", "KIV", "ZAG", "LJU", "BUD", "BTS", "WAW", "KRK", "GDN", "VNO", "RIX", "TLL", "REK", "MOW", "LED", "MSQ", "IEV/KBP", "SJJ", "THR", "ABD", "KBL", "KWI", "RUH", "JED", "DMM", "SAH", "ADE", "BGW", "BEY", "BAH", "AUH", "DXB", "SHJ", "DOH", "JRD", "TLV", "DAM", "AMM", "ANK", "ADA", "BTZ", "IZM", "IST", "BAH", "NIC", "LCA", "BAK", "EVN", "TBS", "MSH", "ASB", "DYU", "KGF", "FRU", "TAS", "CAI", "KRT", "MCT", "ADD", "JIB", "NBO", "TIP", "ALG", "AAE", "TUN", "RBA", "CAS", "NDJ", "NIM", "ABV", "LOS", "PHC", "BKO", "OUA", "COO", "LFW", "ACC", "ASK", "ABJ", "HGS", "MLW", "CKF", "DKR", "BJL", "KLA", "BGF", "YAO", "SSG", "KLA", "KGL", "DAR", "BJM", "BZV", "LBV", "TMS", "MPM", "LLW", "LUN", "HRE", "LAD", "GBE", "WDH", "JNB", "DUR", "CPT", "MRU", "TNR", "YVA", "SEZ", "NKC", "HKG", "TPE", "KHH", "FNJ", "SEL/ICN", "PUS", "TYONRT", "KIX/OSA", "NGO", "FUK", "YOK", "HIJ", "OKA", "SDJ", "SPA", "MNL", "HEB", "DVO", "KUL", "PEN", "LGK", "BKI", "KCH", "IPH", "JHB", "KBR", "SBW", "SDK", "BWN", "SIN", "JKT", "MES", "SUB", "DPS", "UPG", "PNK", "DIL", "SGN", "HAN", "HPH", "VTE", "BKK", "CEI", "HDY", "HKT", "NSI", "RGN", "MDL", "PNH", "DAC", "CGP", "DEL", "BOM", "CCU", "MAA", "BLR", "SXM", "HYD", "KTM", "ISB", "KHI", "LHE", "PEW", "CMB", "MLE", "ULN", "CBR", "MEL", "ADL", "DRN", "CNS", "BNE", "PER", "SYD", "WLG", "AKL", "CHC", "POM", "SUV", "TRW", "HIR", "TBU", "APW", "FUN", "KSA", "VLI"}
	i, v := 0, ""
	for ; i < len(code); i++ {
		if code[i] >= '0' && code[i] <= '9' {
			break
		}
	}
	code = strings.ToUpper(code[:i])
	for i, v = range airPortCode {
		if strings.Contains(code, v) {
			return v
		}
	}
	return code
}
