package model

var Version = "v0.0.1"
var LocalServers = []string{"25", "465", "110", "995", "143", "993"}
var Platforms = []string{"QQ", "163", "Sohu", "Yandex", "Gmail", "Outlook", "Office365", "Yahoo", "MailCOM", "MailRU", "AOL", "GMX", "Sina"}
var SmtpServers = map[string]string{
	"Gmail":     "smtp.gmail.com",
	"163":       "smtp.163.com",
	"Yandex":    "smtp.yandex.com",
	"Office365": "smtp.office365.com",
	"QQ":        "smtp.qq.com",
	"Outlook":   "smtp.outlook.com",
	"Yahoo":     "smtp.mail.yahoo.com",
	// "Apple":     "smtp.mail.me.com",
	"MailRU":  "smtp.mail.ru",
	"AOL":     "smtp.aol.com",
	"GMX":     "smtp.gmx.com",
	"MailCOM": "smtp.mail.com",
	"Sohu":    "smtp.sohu.com",
	"Sina":    "smtp.sina.com",
}
var Pop3Servers = map[string]string{
	"Gmail":     "pop.gmail.com",
	"163":       "pop.163.com",
	"Yandex":    "pop.yandex.com",
	"Office365": "outlook.office365.com",
	"QQ":        "pop.qq.com",
	"Outlook":   "pop-mail.outlook.com",
	"Yahoo":     "pop.mail.yahoo.com",
	// "Apple":     "pop.mail.me.com",
	"MailRU":  "pop.mail.ru",
	"AOL":     "pop.aol.com",
	"GMX":     "pop.gmx.com",
	"MailCOM": "pop.mail.com",
	"Sohu":    "pop.sohu.com",
	"Sina":    "pop.sina.com",
}
var ImapServers = map[string]string{
	"Gmail":     "imap.gmail.com",
	"163":       "imap.163.com",
	"Yandex":    "imap.yandex.com",
	"Office365": "outlook.office365.com",
	"QQ":        "imap.qq.com",
	"Outlook":   "imap-mail.outlook.com",
	"Yahoo":     "imap.mail.yahoo.com",
	// "Apple":     "imap.mail.me.com",
	"MailRU":  "imap.mail.ru",
	"AOL":     "imap.aol.com",
	"GMX":     "imap.gmx.com",
	"MailCOM": "imap.mail.com",
	"Sohu":    "imap.sohu.com",
	"Sina":    "imap.sina.com",
}
