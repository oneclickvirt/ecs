package model

const UnlockTestsVersion = "v0.0.12"

var EnableLoger = false

type Result struct {
	Name       string
	Status     string
	Err        error
	Region     string
	Info       string
	UnlockType string
}

const (
	StatusUnexpected = "Unknown"
	StatusNetworkErr = "NetworkError"
	StatusErr        = "Error"
	StatusRestricted = "Restricted"
	StatusYes        = "Yes"
	StatusNo         = "No"
	StatusBanned     = "Banned"
	PrintHead        = "PrintHead"
	UA_Browser       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	UA_SecCHUA       = "\"Chromium\";v=\"124\", \"Google Chrome\";v=\"124\", \"Not-A.Brand\";v=\"99\""
	UA_Dalvik        = "Mozilla/5.0 (Linux; Android 10; Pixel 4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Mobile Safari/537.36"
	UA_Pjsekai       = "pjsekai/48 CFNetwork/1240.0.4 Darwin/20.6.0"
)

var StarPlusSupportCountry = []string{
	"br", "mx", "ar", "cl", "co", "pe", "uy", "ec", "pa", "cr", "py", "bo",
	"gt", "ni", "do", "sv", "hn", "ve",
}

var GptSupportCountry = []string{
	"al", "dz", "ad", "ao", "ag", "ar", "am", "au", "at", "az", "bs", "bd",
	"bb", "be", "bz", "bj", "bt", "ba", "bw", "br", "bg", "bf", "cv", "ca",
	"cl", "co", "km", "cr", "hr", "cy", "dk", "dj", "dm", "do", "ec", "sv",
	"ee", "fj", "fi", "fr", "ga", "gm", "ge", "de", "gh", "gr", "gd", "gt",
	"gn", "gw", "gy", "ht", "hn", "hu", "is", "in", "id", "iq", "ie", "il",
	"it", "jm", "jp", "jo", "kz", "ke", "ki", "kw", "kg", "lv", "lb", "ls",
	"lr", "li", "lt", "lu", "mg", "mw", "my", "mv", "ml", "mt", "mh", "mr",
	"mu", "mx", "mc", "mn", "me", "ma", "mz", "mm", "na", "nr", "np", "nl",
	"nz", "ni", "ne", "ng", "mk", "no", "om", "pk", "pw", "pa", "pg", "pe",
	"ph", "pl", "pt", "qa", "ro", "rw", "kn", "lc", "vc", "ws", "sm", "st",
	"sn", "rs", "sc", "sl", "sg", "sk", "si", "sb", "za", "es", "lk", "sr",
	"se", "ch", "th", "tg", "to", "tt", "tn", "tr", "tv", "ug", "ae", "us",
	"uy", "vu", "zm", "bo", "bn", "cg", "cz", "va", "fm", "md", "ps", "kr",
	"tw", "tz", "tl", "gb",
}

var DiscoveryPlusSupportCountry = []string{
	"at", "br", "ca", "dk", "fi", "de", "in", "ie", "it", "nl", "no", "es",
	"se", "gb", "us"}

var NLZIETSupportCountry = []string{
	"be", "bg", "cz", "dk", "de", "ee", "ie", "el", "es", "fr",
	"hr", "it", "cy", "lv", "lt", "lu", "hu", "mt", "nl", "at",
	"pl", "pt", "ro", "si", "sk", "fi", "se",
}
