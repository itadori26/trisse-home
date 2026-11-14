package gateway

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/modules/parcels/usecase"
)

// IMAP reads shipping mails from one folder (#166). It is deliberately
// minimal and read-only: the folder is opened with EXAMINE (the server
// refuses any change), bodies are fetched with BODY.PEEK (not even the
// \Seen flag changes), and only headers and the first part of the text are
// read; attachments are ignored. Implicit TLS (port 993) only.
type IMAP struct {
	dial DialFunc
	// plain skips TLS; tests only.
	plain bool
}

// DialFunc opens a TCP connection; production passes the guarded dialer of
// the integration clients, so link-local and metadata addresses are refused.
type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

// NewIMAP returns the mail source.
func NewIMAP(dial DialFunc) *IMAP { return &IMAP{dial: dial} }

// Limits of one run.
const (
	imapTimeout     = 60 * time.Second
	maxMailsPerRun  = 50
	maxTextBytes    = 256 << 10
	maxLiteralBytes = maxTextBytes + 64<<10
	maxLineBytes    = 64 << 10
)

// Test implements usecase.MailSource.
func (m *IMAP) Test(ctx context.Context, s domain.MailSettings) error {
	c, err := m.open(ctx, s)
	if err != nil {
		return err
	}
	defer c.close()
	_, _, err = c.examine(s.Folder)
	return err
}

// Fetch implements usecase.MailSource.
func (m *IMAP) Fetch(ctx context.Context, s domain.MailSettings, st usecase.MailState, since time.Time) ([]domain.Mail, usecase.MailState, error) {
	c, err := m.open(ctx, s)
	if err != nil {
		return nil, st, err
	}
	defer c.close()
	validity, uidNext, err := c.examine(s.Folder)
	if err != nil {
		return nil, st, err
	}
	next := usecase.MailState{UIDValidity: validity, LastUID: st.LastUID}
	var search string
	if st.UIDValidity != validity || st.LastUID == 0 {
		next.LastUID = 0
		search = "UID SEARCH SINCE " + since.Format("2-Jan-2006")
	} else {
		search = fmt.Sprintf("UID SEARCH UID %d:*", st.LastUID+1)
	}
	lines, err := c.cmd(search)
	if err != nil {
		return nil, st, err
	}
	var uids []uint32
	for _, l := range lines {
		if rest, ok := strings.CutPrefix(l.text, "* SEARCH"); ok {
			for _, f := range strings.Fields(rest) {
				// "n:*" always returns the last message, even an old one.
				if u, err := strconv.ParseUint(f, 10, 32); err == nil && uint32(u) > next.LastUID {
					uids = append(uids, uint32(u))
				}
			}
		}
	}
	slices.Sort(uids)
	if len(uids) > maxMailsPerRun {
		uids = uids[len(uids)-maxMailsPerRun:] // the newest
	}
	if uidNext > 0 && uidNext-1 > next.LastUID {
		next.LastUID = uidNext - 1
	}
	if len(uids) == 0 {
		return nil, next, nil
	}
	next.LastUID = max(next.LastUID, uids[len(uids)-1])
	set := make([]string, len(uids))
	for i, u := range uids {
		set[i] = strconv.FormatUint(uint64(u), 10)
	}
	lines, err = c.cmd("UID FETCH " + strings.Join(set, ",") +
		" (UID BODY.PEEK[HEADER.FIELDS (FROM SUBJECT CONTENT-TYPE CONTENT-TRANSFER-ENCODING)] BODY.PEEK[TEXT]<0." + strconv.Itoa(maxTextBytes) + ">)")
	if err != nil {
		return nil, st, err
	}
	var mails []domain.Mail
	for _, l := range lines {
		if !fetchLine.MatchString(l.text) {
			continue
		}
		items := l.items()
		header, text := items["HEADER"], items["TEXT"]
		if header == nil {
			continue
		}
		mails = append(mails, ParseMail(header, text))
	}
	return mails, next, nil
}

var (
	fetchLine   = regexp.MustCompile(`^\* \d+ FETCH `)
	literalTail = regexp.MustCompile(`\{(\d+)\}$`)
	uidValidity = regexp.MustCompile(`\[UIDVALIDITY (\d+)\]`)
	uidNextRe   = regexp.MustCompile(`\[UIDNEXT (\d+)\]`)
	sectionRe   = regexp.MustCompile(`BODY\[(HEADER|TEXT)[^\]]*\](?:<\d+>)? `)
)

type conn struct {
	nc  net.Conn
	r   *bufio.Reader
	tag int
}

// response is one server response; literals are cut out of the text, at
// marks the position in text where each literal was.
type response struct {
	text string
	lits [][]byte
	at   []int
}

func (m *IMAP) open(ctx context.Context, s domain.MailSettings) (*conn, error) {
	ctx, cancel := context.WithTimeout(ctx, imapTimeout)
	defer cancel()
	nc, err := m.dial(ctx, "tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.Port)))
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if !m.plain {
		tc := tls.Client(nc, &tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12})
		if err := tc.HandshakeContext(ctx); err != nil {
			_ = nc.Close()
			return nil, fmt.Errorf("TLS (only IMAP over TLS, usually port 993, is supported): %w", err)
		}
		nc = tc
	}
	_ = nc.SetDeadline(time.Now().Add(imapTimeout))
	c := &conn{nc: nc, r: bufio.NewReaderSize(nc, 16<<10)}
	greeting, err := c.read()
	if err != nil || !strings.HasPrefix(greeting.text, "* OK") {
		c.close()
		return nil, fmt.Errorf("not an IMAP server: %q %w", clipText(greeting.text), err)
	}
	if _, err := c.cmd("LOGIN " + quote(s.Username) + " " + quote(s.Password)); err != nil {
		c.close()
		return nil, errors.New("login failed: check user name and app password")
	}
	return c, nil
}

func (c *conn) close() {
	_, _ = c.cmd("LOGOUT")
	_ = c.nc.Close()
}

// examine opens the folder read-only and returns UIDVALIDITY and UIDNEXT.
func (c *conn) examine(folder string) (uint32, uint32, error) {
	lines, err := c.cmd("EXAMINE " + quote(folder))
	if err != nil {
		return 0, 0, fmt.Errorf("folder %q: %w", folder, err)
	}
	var validity, next uint32
	for _, l := range lines {
		if m := uidValidity.FindStringSubmatch(l.text); m != nil {
			v, _ := strconv.ParseUint(m[1], 10, 32)
			validity = uint32(v)
		}
		if m := uidNextRe.FindStringSubmatch(l.text); m != nil {
			v, _ := strconv.ParseUint(m[1], 10, 32)
			next = uint32(v)
		}
	}
	return validity, next, nil
}

// cmd sends a command and returns the untagged responses.
func (c *conn) cmd(command string) ([]response, error) {
	c.tag++
	tag := "T" + strconv.Itoa(c.tag)
	if _, err := io.WriteString(c.nc, tag+" "+command+"\r\n"); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}
	var out []response
	for {
		r, err := c.read()
		if err != nil {
			return nil, err
		}
		if rest, ok := strings.CutPrefix(r.text, tag+" "); ok {
			if strings.HasPrefix(rest, "OK") {
				return out, nil
			}
			return nil, fmt.Errorf("server answered: %s", clipText(rest))
		}
		out = append(out, r)
	}
}

// read reads one response including its literals.
func (c *conn) read() (response, error) {
	var r response
	for {
		line, err := c.line()
		if err != nil {
			return r, err
		}
		m := literalTail.FindStringSubmatchIndex(line)
		if m == nil {
			r.text += line
			return r, nil
		}
		n, _ := strconv.Atoi(line[m[2]:m[3]])
		if n > maxLiteralBytes {
			return r, fmt.Errorf("message part too large (%d bytes)", n)
		}
		r.text += line[:m[0]]
		buf := make([]byte, n)
		if _, err := io.ReadFull(c.r, buf); err != nil {
			return r, fmt.Errorf("read literal: %w", err)
		}
		r.lits = append(r.lits, buf)
		r.at = append(r.at, len(r.text))
	}
}

func (c *conn) line() (string, error) {
	var b strings.Builder
	for {
		chunk, err := c.r.ReadSlice('\n')
		b.Write(chunk)
		if b.Len() > maxLineBytes {
			return "", errors.New("response line too long")
		}
		if err == nil {
			return strings.TrimRight(b.String(), "\r\n"), nil
		}
		if !errors.Is(err, bufio.ErrBufferFull) {
			return "", fmt.Errorf("read: %w", err)
		}
	}
}

// items returns the HEADER and TEXT sections of a FETCH response, sent as
// literal, quoted string or NIL.
func (r response) items() map[string][]byte {
	out := map[string][]byte{}
	for _, m := range sectionRe.FindAllStringSubmatchIndex(r.text, -1) {
		name, pos := r.text[m[2]:m[3]], m[1]
		if i := slices.Index(r.at, pos); i >= 0 {
			out[name] = r.lits[i]
			continue
		}
		if pos < len(r.text) && r.text[pos] == '"' {
			out[name] = []byte(unquote(r.text[pos:]))
		}
	}
	return out
}

func quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

// unquote reads a quoted string at the start of s.
func unquote(s string) string {
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if i+1 < len(s) {
				i++
				b.WriteByte(s[i])
			}
		case '"':
			return b.String()
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func clipText(s string) string {
	if len(s) > 120 {
		return s[:120] + "…"
	}
	return s
}
