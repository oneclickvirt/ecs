package tests

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	mailmodel "github.com/oneclickvirt/portchecker/model"
)

func TestLegacyMailSilentPeerCannotHoldTheWorkflow(t *testing.T) {
	for _, secure := range []bool{false, true} {
		t.Run(fmt.Sprint(secure), func(t *testing.T) {
			client, server := net.Pipe()
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()
			done := make(chan bool, 1)
			go func() {
				done <- probeLegacyMail(ctx, "fixture.test", "25", secure,
					func(context.Context, string, string) (net.Conn, error) { return client, nil })
			}()
			select {
			case ok := <-done:
				if ok {
					t.Fatal("silent peer was marked available")
				}
			case <-time.After(time.Second):
				t.Fatal("silent greeting or TLS handshake blocked past deadline")
			}
		})
	}
}

func TestLegacyMailCancellationClosesEstablishedConnection(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dialed, done := make(chan struct{}), make(chan bool, 1)
	go func() {
		done <- probeLegacyMail(ctx, "fixture.test", "110", false,
			func(context.Context, string, string) (net.Conn, error) { close(dialed); return client, nil })
	}()
	<-dialed
	cancel()
	select {
	case ok := <-done:
		if ok {
			t.Fatal("canceled probe succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not interrupt the greeting read")
	}
}

func TestLegacyMailGreetingAndTableFormat(t *testing.T) {
	for _, greeting := range []string{"220 smtp ready\r\n", "+OK POP3 ready\r\n", "* OK IMAP ready\r\n"} {
		client, server := net.Pipe()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		go func() { defer server.Close(); _, _ = fmt.Fprint(server, greeting) }()
		ok := probeLegacyMail(ctx, "fixture.test", "25", false,
			func(context.Context, string, string) (net.Conn, error) { return client, nil })
		cancel()
		if !ok {
			t.Fatalf("valid greeting rejected: %q", greeting)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	lines := strings.Split(EmailCheckContext(ctx), "\n")
	wantHeader := fmt.Sprintf("%-9s %-5s %-5s %-5s %-5s %-5s %-5s", "Platform", "SMTP", "SMTPS", "POP3", "POP3S", "IMAP", "IMAPS")
	if lines[0] != wantHeader || len(lines) != len(mailmodel.Platforms)+2 {
		t.Fatalf("classic mail table changed: %q", lines)
	}
	for index, name := range append([]string{"LocalPort"}, mailmodel.Platforms...) {
		if fields := strings.Fields(strings.TrimPrefix(lines[index+1], name)); len(fields) != 6 || !strings.HasPrefix(lines[index+1], name) {
			t.Fatalf("classic row order or columns changed: %q", lines[index+1])
		}
	}
}
