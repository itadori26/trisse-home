package gateway

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/notifications/domain"
)

var msg = domain.Message{Title: `Disk "/" 95 %`, Body: "line1\nline2", Severity: domain.SeverityCritical, Type: "disk"}

func TestWebhookAndGotify(t *testing.T) {
	var got []byte
	var sig, key string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = io.ReadAll(r.Body)
		sig, key = r.Header.Get("X-Trisse-Signature"), r.Header.Get("X-Gotify-Key")
	}))
	defer srv.Close()
	hook := domain.Channel{Kind: domain.ChannelWebhook, Name: "h", URL: srv.URL, Secret: "s3cret"}
	_ = hook.Validate()
	if err := NewWebhookPublisher(srv.Client()).Publish(context.Background(), hook, msg); err != nil {
		t.Fatal(err)
	}
	var body map[string]string
	if err := json.Unmarshal(got, &body); err != nil || body["title"] != msg.Title || body["severity"] != "critical" {
		t.Fatalf("body %s: %v", got, err)
	}
	if sig != Signature("s3cret", got) {
		t.Fatal("signature missing or wrong")
	}
	if err := NewGotifyPublisher(srv.Client()).Publish(context.Background(), domain.Channel{URL: srv.URL + "/", Secret: "app"}, msg); err != nil || key != "app" {
		t.Fatalf("gotify: %v %q", err, key)
	}
}

// fakeSMTP speaks just enough SMTP for a plain (TLS "none") delivery.
func fakeSMTP(t *testing.T, conn net.Conn, data *strings.Builder) {
	t.Helper()
	defer conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	reply := func(s string) { _, _ = rw.WriteString(s + "\r\n"); _ = rw.Flush() }
	reply("220 test")
	inData := false
	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		if inData {
			if line == ".\r\n" {
				inData = false
				reply("250 queued")
				continue
			}
			data.WriteString(line)
			continue
		}
		switch cmd := strings.ToUpper(strings.TrimSpace(line)); {
		case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
			reply("250 test")
		case strings.HasPrefix(cmd, "MAIL"), strings.HasPrefix(cmd, "RCPT"):
			reply("250 ok")
		case cmd == "DATA":
			inData = true
			reply("354 go")
		case cmd == "QUIT":
			reply("221 bye")
			return
		default:
			reply("502 no")
		}
	}
}

func TestEmail(t *testing.T) {
	client, server := net.Pipe()
	var data strings.Builder
	done := make(chan struct{})
	go func() { fakeSMTP(t, server, &data); close(done) }()
	p := NewEmailPublisher(func() time.Time { return time.Unix(0, 0) })
	p.Dial = func(context.Context, string) (net.Conn, error) { return client, nil }
	c := domain.Channel{Kind: domain.ChannelEmail, Name: "m", SMTPHost: "mail.example.org", From: "home@example.org", To: "a@example.org, b@example.org", TLS: "none"}
	_ = c.Validate()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Publish(ctx, c, msg); err != nil {
		t.Fatal(err)
	}
	<-done
	if !strings.Contains(data.String(), "Subject: [trisse-home] Disk") || !strings.Contains(data.String(), "line2") {
		t.Fatalf("mail:\n%s", data.String())
	}
	umlaut := string(BuildEmail(c, domain.Message{Title: "Größe"}, time.Unix(0, 0)))
	if !strings.Contains(umlaut, "Subject: =?utf-8?q?") {
		t.Fatalf("non-ASCII subject not encoded: %s", umlaut)
	}
}
