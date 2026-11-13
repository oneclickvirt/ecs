package tests

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	mailmodel "github.com/oneclickvirt/portchecker/model"
)

// EmailCheckContext keeps portchecker's classic six-column table and target
// registry. Bound the whole connection, TLS handshake and greeting read: the
// legacy EmailCheck only bounds dialing, so a silent peer can block forever.
func EmailCheckContext(ctx context.Context) string {
	ports := []string{"25", "465", "110", "995", "143", "993"}
	rows := make([][6]string, len(mailmodel.Platforms)+1)
	var wg sync.WaitGroup
	for column, port := range ports {
		wg.Go(func() {
			listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", ":"+port)
			rows[0][column] = legacyMailMark(err == nil)
			if err == nil {
				_ = listener.Close()
			}
		})
	}
	for index, platform := range mailmodel.Platforms {
		hosts := []string{
			mailmodel.SmtpServers[platform], mailmodel.SmtpServers[platform],
			mailmodel.Pop3Servers[platform], mailmodel.Pop3Servers[platform],
			mailmodel.ImapServers[platform], mailmodel.ImapServers[platform],
		}
		for column, port := range ports {
			wg.Go(func() {
				probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				ok := probeLegacyMail(probeCtx, hosts[column], port, column%2 == 1, (&net.Dialer{}).DialContext)
				rows[index+1][column] = legacyMailMark(ok)
			})
		}
	}
	wg.Wait()
	lines := []string{fmt.Sprintf("%-9s %-5s %-5s %-5s %-5s %-5s %-5s", "Platform", "SMTP", "SMTPS", "POP3", "POP3S", "IMAP", "IMAPS")}
	for index, row := range rows {
		name := "LocalPort"
		if index > 0 {
			name = mailmodel.Platforms[index-1]
		}
		lines = append(lines, fmt.Sprintf("%-10s%-5s %-5s %-5s %-5s %-5s %-5s", name, row[0], row[1], row[2], row[3], row[4], row[5]))
	}
	return strings.Join(lines, "\n")
}

func probeLegacyMail(ctx context.Context, host, port string, secure bool, dial func(context.Context, string, string) (net.Conn, error)) bool {
	if host == "" || ctx.Err() != nil {
		return false
	}
	conn, err := dial(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return false
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return false
		}
	}
	var reader io.Reader = conn
	if secure {
		tlsConn := tls.Client(conn, &tls.Config{ServerName: host})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return false
		}
		reader = tlsConn
	}
	response, err := bufio.NewReader(io.LimitReader(reader, 4096)).ReadString('\n')
	if err != nil {
		return false
	}
	// SMTP, POP3 and IMAP use the same greeting after TLS as in plaintext.
	return strings.HasPrefix(response, "220 ") || strings.HasPrefix(response, "220-") ||
		strings.HasPrefix(response, "+OK") || strings.HasPrefix(response, "* OK")
}

func legacyMailMark(ok bool) string {
	if runtime.GOOS == "windows" {
		if ok {
			return "Y"
		}
		return "N"
	}
	if ok {
		return "✔"
	}
	return "✘"
}
