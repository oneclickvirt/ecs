package email

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/oneclickvirt/portchecker/model"
)

type Result struct {
	Platform string
	Status   string
}

func isLocalPortOpen(port string) string {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return "✘"
	}
	ln.Close()
	return "✔"
}

func checkConnection(host string, port string) string {
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return "✘"
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "✘"
	}
	status := "✘"
	if strings.Split(response, " ")[0] == "220" || strings.Split(response, " ")[0] == "+OK" || strings.Contains(response, "* OK") {
		status = "✔"
	}
	return status
}

func checkConnectionSSL(host string, port string) string {
	address := net.JoinHostPort(host, port)
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", address, &tls.Config{})
	if err != nil {
		return "✘"
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "✘"
	}
	status := "✘"
	if strings.HasPrefix(response, "220") {
		status = "✔"
	}
	return status
}

func EmailCheck() string {
	var wg sync.WaitGroup
	localChan := make(chan Result, len(model.LocalServers))
	smtpChan := make(chan Result, len(model.SmtpServers))
	pop3Chan := make(chan Result, len(model.Pop3Servers))
	imapChan := make(chan Result, len(model.ImapServers))
	smtpsChan := make(chan Result, len(model.SmtpServers))
	pop3sChan := make(chan Result, len(model.Pop3Servers))
	imapsChan := make(chan Result, len(model.ImapServers))
	checkLocal := func(port string) {
		defer wg.Done()
		localResult := isLocalPortOpen(port)
		localChan <- Result{port, localResult}
	}
	checkSMTP := func(name, host string) {
		defer wg.Done()
		smtpResult := checkConnection(host, "25")
		smtpChan <- Result{name, smtpResult}
	}
	checkSMTPS := func(name, host string) {
		defer wg.Done()
		smtpSSLResult := checkConnectionSSL(host, "465")
		smtpsChan <- Result{name, smtpSSLResult}
	}
	checkPOP3 := func(name, host string) {
		defer wg.Done()
		pop3Result := checkConnection(host, "110")
		pop3Chan <- Result{name, pop3Result}
	}
	checkPOP3S := func(name, host string) {
		defer wg.Done()
		pop3SSLResult := checkConnectionSSL(host, "995")
		pop3sChan <- Result{name, pop3SSLResult}
	}
	checkIMAP := func(name, host string) {
		defer wg.Done()
		imapResult := checkConnection(host, "143")
		imapChan <- Result{name, imapResult}
	}
	checkIMAPS := func(name, host string) {
		defer wg.Done()
		imapSSLResult := checkConnectionSSL(host, "993")
		imapsChan <- Result{name, imapSSLResult}
	}
	for _, port := range model.LocalServers {
		wg.Add(1)
		go checkLocal(port)
	}
	for name, smtpHost := range model.SmtpServers {
		wg.Add(1)
		go checkSMTP(name, smtpHost)
	}
	for name, smtpHost := range model.SmtpServers {
		wg.Add(1)
		go checkSMTPS(name, smtpHost)
	}
	for name, pop3Host := range model.Pop3Servers {
		wg.Add(1)
		go checkPOP3(name, pop3Host)
	}
	for name, pop3Host := range model.Pop3Servers {
		wg.Add(1)
		go checkPOP3S(name, pop3Host)
	}
	for name, imapHost := range model.ImapServers {
		wg.Add(1)
		go checkIMAP(name, imapHost)
	}
	for name, imapHost := range model.ImapServers {
		wg.Add(1)
		go checkIMAPS(name, imapHost)
	}
	wg.Wait()
	close(localChan)
	close(smtpChan)
	close(pop3Chan)
	close(imapChan)
	close(smtpsChan)
	close(pop3sChan)
	close(imapsChan)
	//转换通道提取数据
	temp := []string{}
	smtpChanMap, smtpsChanMap, pop3ChanMap, pop3sChanMap, imapChanMap, imapsChanMap := map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}
	wg.Add(7)
	// 使用goroutine并发处理每个通道
	go func() {
		defer wg.Done()
		for m := range localChan {
			temp = append(temp, m.Status)
		}
	}()
	go func() {
		defer wg.Done()
		for m := range smtpChan {
			smtpChanMap[m.Platform] = m.Status
		}
	}()
	go func() {
		defer wg.Done()
		for m := range smtpsChan {
			smtpsChanMap[m.Platform] = m.Status
		}
	}()
	go func() {
		defer wg.Done()
		for m := range pop3Chan {
			pop3ChanMap[m.Platform] = m.Status
		}
	}()
	go func() {
		defer wg.Done()
		for m := range pop3sChan {
			pop3sChanMap[m.Platform] = m.Status
		}
	}()
	go func() {
		defer wg.Done()
		for m := range imapChan {
			imapChanMap[m.Platform] = m.Status
		}
	}()
	go func() {
		defer wg.Done()
		for m := range imapsChan {
			imapsChanMap[m.Platform] = m.Status
		}
	}()
	wg.Wait()
	var results []string
	results = append(results, fmt.Sprintf("%-9s %-5s %-5s %-5s %-5s %-5s %-5s", "Platform", "SMTP", "SMTPS", "POP3", "POP3S", "IMAP", "IMAPS"))
	results = append(results, fmt.Sprintf("%-10s%-5s %-5s %-5s %-5s %-5s %-5s", "LocalPort", temp[0], temp[1], temp[2], temp[3], temp[4], temp[5]))
	for _, name := range model.Platforms {
		results = append(results, fmt.Sprintf("%-10s%-5s %-5s %-5s %-5s %-5s %-5s", name,
			smtpChanMap[name], smtpsChanMap[name], pop3ChanMap[name], pop3sChanMap[name], imapChanMap[name], imapsChanMap[name]))
	}
	return strings.Join(results, "\n")
}