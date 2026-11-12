package gateway

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/notifications/domain"
)

func post(ctx context.Context, client *http.Client, url string, body []byte, headers map[string]string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64<<10))
	if res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("answered %s", res.Status)
	}
	return nil
}

// GotifyPublisher delivers to a Gotify server.
type GotifyPublisher struct{ client *http.Client }

// NewGotifyPublisher returns a publisher using client.
func NewGotifyPublisher(client *http.Client) *GotifyPublisher {
	return &GotifyPublisher{client: client}
}

// gotifyPriority maps severities to Gotify's 0–10 scale.
func gotifyPriority(s domain.Severity) int {
	switch s {
	case domain.SeverityCritical:
		return 8
	case domain.SeverityWarning:
		return 5
	}
	return 2
}

// Publish implements usecase.Publisher.
func (g *GotifyPublisher) Publish(ctx context.Context, c domain.Channel, m domain.Message) error {
	body, _ := json.Marshal(map[string]any{"title": m.Title, "message": m.Body, "priority": gotifyPriority(m.Severity)}) // plain map: cannot fail
	if err := post(ctx, g.client, strings.TrimRight(c.URL, "/")+"/message", body, map[string]string{"X-Gotify-Key": c.Secret}); err != nil {
		return fmt.Errorf("gotify %w", err)
	}
	return nil
}

// WebhookPublisher posts a JSON body to any URL.
type WebhookPublisher struct{ client *http.Client }

// NewWebhookPublisher returns a publisher using client.
func NewWebhookPublisher(client *http.Client) *WebhookPublisher {
	return &WebhookPublisher{client: client}
}

// RenderWebhook fills the template; values are inserted JSON-encoded, so
// message text can never break the JSON structure.
func RenderWebhook(template string, m domain.Message) []byte {
	enc := func(s string) string {
		b, _ := json.Marshal(s) // a string: cannot fail
		return string(b)
	}
	r := strings.NewReplacer("{{title}}", enc(m.Title), "{{body}}", enc(m.Body), "{{severity}}", enc(string(m.Severity)), "{{type}}", enc(m.Type))
	return []byte(r.Replace(template))
}

// Signature returns the HMAC-SHA256 of body as "sha256=<hex>".
func Signature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// Publish implements usecase.Publisher.
func (w *WebhookPublisher) Publish(ctx context.Context, c domain.Channel, m domain.Message) error {
	body := RenderWebhook(c.Template, m)
	headers := map[string]string{}
	if c.Secret != "" {
		headers["X-Trisse-Signature"] = Signature(c.Secret, body)
	}
	if err := post(ctx, w.client, c.URL, body, headers); err != nil {
		return fmt.Errorf("webhook %w", err)
	}
	return nil
}

// EmailPublisher sends e-mails via SMTP.
type EmailPublisher struct {
	// Dial is replaceable in tests.
	Dial func(ctx context.Context, addr string) (net.Conn, error)
	now  func() time.Time
}

// NewEmailPublisher returns an SMTP publisher.
func NewEmailPublisher(now func() time.Time) *EmailPublisher {
	var d net.Dialer
	return &EmailPublisher{Dial: func(ctx context.Context, addr string) (net.Conn, error) { return d.DialContext(ctx, "tcp", addr) }, now: now}
}

// headerValue removes control characters (CR, LF, …) so values taken from
// service names or alert titles cannot start new header lines.
func headerValue(s string) string {
	return strings.Map(func(r rune) rune {
		if r < ' ' || r == 0x7f {
			return ' '
		}
		return r
	}, s)
}

// bodyText normalises line endings to CRLF and drops other control
// characters; dot-stuffing is done by the SMTP data writer.
func bodyText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.Map(func(r rune) rune {
		if (r < ' ' && r != '\n' && r != '\t') || r == 0x7f {
			return -1
		}
		return r
	}, s)
	return strings.ReplaceAll(s, "\n", "\r\n")
}

// BuildEmail renders the message as an RFC 5322 mail with a UTF-8 subject.
func BuildEmail(c domain.Channel, m domain.Message, date time.Time) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\nTo: %s\r\n", headerValue(c.From), headerValue(c.To))
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", headerValue("[trisse-home] "+m.Title)))
	fmt.Fprintf(&b, "Date: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n", date.Format(time.RFC1123Z))
	fmt.Fprintf(&b, "%s\r\n\r\nSeverity: %s\r\n", bodyText(m.Body), headerValue(string(m.Severity)))
	return []byte(b.String())
}

// Publish implements usecase.Publisher.
func (e *EmailPublisher) Publish(ctx context.Context, c domain.Channel, m domain.Message) error {
	addr := net.JoinHostPort(c.SMTPHost, strconv.Itoa(c.SMTPPort))
	conn, err := e.Dial(ctx, addr)
	if err != nil {
		return fmt.Errorf("smtp connect: %w", err)
	}
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(dl)
	}
	tlsConfig := &tls.Config{ServerName: c.SMTPHost, MinVersion: tls.VersionTLS12}
	if c.TLS == "tls" {
		conn = tls.Client(conn, tlsConfig)
	}
	client, err := smtp.NewClient(conn, c.SMTPHost)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("smtp handshake: %w", err)
	}
	defer client.Close()
	if c.TLS == "starttls" {
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}
	if c.Username != "" {
		// PlainAuth refuses to send the password without TLS (except localhost).
		if err := client.Auth(smtp.PlainAuth("", c.Username, c.Secret, c.SMTPHost)); err != nil {
			return fmt.Errorf("smtp login: %w", err)
		}
	}
	if err := client.Mail(c.From); err != nil {
		return fmt.Errorf("smtp sender: %w", err)
	}
	for _, to := range strings.Split(c.To, ",") {
		if err := client.Rcpt(strings.TrimSpace(to)); err != nil {
			return fmt.Errorf("smtp recipient: %w", err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(BuildEmail(c, m, e.now())); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return client.Quit()
}
