package printhead

import "fmt"

// PrintDatabaseInfo 打印数据库头部信息
func PrintDatabaseInfo(language string) (string, error) {
	var res string
	if language == "en" {
		res += "The following is the number of each database, the output will come with the corresponding number of the database source\n"
		res += "ipinfo  databases [0] | scamalytics databases [1] | virustotal  databases  [2] | abuseipdb databases   [3] | ip2location databases    [4]\n"
		res += "ip-api  databases [5] | ipwhois databases     [6] | ipregistry  databases  [7] | ipdata databases      [8] | db-ip databases          [9]\n"
		res += "ipapiis databases [A] | ipapicom databases    [B] | bigdatacloud databases [C] | cheervision databases [D] | ipqualityscore databases [E]\n"
		return res, nil
	} else if language == "zh" {
		res += "以下为各数据库编号，输出结果后将自带数据库来源对应的编号\n"
		res += "ipinfo数据库  [0] | scamalytics数据库 [1] | virustotal数据库   [2] | abuseipdb数据库   [3] | ip2location数据库    [4]\n"
		res += "ip-api数据库  [5] | ipwhois数据库     [6] | ipregistry数据库   [7] | ipdata数据库      [8] | db-ip数据库          [9]\n"
		res += "ipapiis数据库 [A] | ipapicom数据库    [B] | bigdatacloud数据库 [C] | cheervision数据库 [D] | ipqualityscore数据库 [E]\n"
		return res, nil
	}
	return "", fmt.Errorf("wrong language")
}
