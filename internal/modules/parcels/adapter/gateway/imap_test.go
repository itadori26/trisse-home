package gateway

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/modules/parcels/usecase"
)

// fakeIMAP is a scripted IMAP server: it records every command and fails
// the test on anything that could change the mailbox.
type fakeIMAP struct {
	t        *testing.T
	ln       net.Listener
	mu       sync.Mutex
	commands []string
	mails    map[uint32]string // uid -> raw message
	validity uint32
}

var writeCommands = regexp.MustCompile(`(?i)^(?:SELECT|STORE|UID STORE|COPY|UID COPY|MOVE|UID MOVE|EXPUNGE|UID EXPUNGE|DELETE|APPEND|RENAME|CREATE)\b`)

func newFakeIMAP(t *testing.T) *fakeIMAP {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeIMAP{t: t, ln: ln, mails: map[uint32]string{}, validity: 7}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go f.serve(c)
		}
	}()
	t.Cleanup(func() { _ = ln.Close() })
	return f
}

func (f *fakeIMAP) serve(c net.Conn) {
	defer c.Close()
	r := bufio.NewReader(c)
	fmt.Fprint(c, "* OK [CAPABILITY IMAP4rev1] ready\r\n")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		tag, cmd, _ := strings.Cut(line, " ")
		f.mu.Lock()
		f.commands = append(f.commands, cmd)
		f.mu.Unlock()
		if writeCommands.MatchString(cmd) || (strings.Contains(cmd, "BODY[") && !strings.Contains(cmd, "BODY.PEEK[")) {
			f.t.Errorf("command could change the mailbox: %s", cmd)
		}
		switch {
		case strings.HasPrefix(cmd, "LOGIN"):
			if cmd != `LOGIN "me@example.com" "app \"pw\""` {
				fmt.Fprintf(c, "%s NO [AUTHENTICATIONFAILED] invalid\r\n", tag)
				continue
			}
			fmt.Fprintf(c, "%s OK logged in\r\n", tag)
		case strings.HasPrefix(cmd, "EXAMINE"):
			if cmd != `EXAMINE "Pakete"` {
				fmt.Fprintf(c, "%s NO no such folder\r\n", tag)
				continue
			}
			fmt.Fprintf(c, "* %d EXISTS\r\n* OK [UIDVALIDITY %d] ok\r\n* OK [UIDNEXT %d] ok\r\n%s OK [READ-ONLY] done\r\n", len(f.mails), f.validity, 20, tag)
		case strings.HasPrefix(cmd, "UID SEARCH"):
			var uids []string
			for u := range f.mails {
				uids = append(uids, fmt.Sprint(u))
			}
			fmt.Fprintf(c, "* SEARCH %s\r\n%s OK done\r\n", strings.Join(uids, " "), tag)
		case strings.HasPrefix(cmd, "UID FETCH"):
			set := strings.Fields(cmd)[2]
			for i, u := range strings.Split(set, ",") {
				var uid uint32
				_, _ = fmt.Sscan(u, &uid)
				raw, ok := f.mails[uid]
				if !ok {
					continue
				}
				head, body, _ := strings.Cut(raw, "\r\n\r\n")
				head += "\r\n\r\n"
				fmt.Fprintf(c, "* %d FETCH (UID %d BODY[HEADER.FIELDS (FROM SUBJECT CONTENT-TYPE CONTENT-TRANSFER-ENCODING)] {%d}\r\n%s BODY[TEXT]<0> {%d}\r\n%s)\r\n",
					i+1, uid, len(head), head, len(body), body)
			}
			fmt.Fprintf(c, "%s OK done\r\n", tag)
		case cmd == "LOGOUT":
			fmt.Fprintf(c, "* BYE\r\n%s OK bye\r\n", tag)
			return
		default:
			fmt.Fprintf(c, "%s BAD unknown\r\n", tag)
		}
	}
}

func (f *fakeIMAP) settings() domain.MailSettings {
	host, port, _ := net.SplitHostPort(f.ln.Addr().String())
	var p int
	_, _ = fmt.Sscan(port, &p)
	return domain.MailSettings{Enabled: true, Host: host, Port: p, Username: "me@example.com", Password: `app "pw"`, Folder: "Pakete", Interval: time.Minute}
}

func plainIMAP() *IMAP {
	return &IMAP{dial: (&net.Dialer{}).DialContext, plain: true}
}

const dhlMail = "From: DHL Paket <noreply@dhl.de>\r\nSubject: =?UTF-8?Q?Ihre_Sendung_von_B=C3=BCcher_Huber?=\r\nContent-Type: multipart/alternative; boundary=\"b1\"\r\n\r\n" +
	"--b1\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\nSendungsnummer: 0034043416109404=\r\n2557\r\n" +
	"--b1\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<p>Hallo</p>\r\n--b1--\r\n"

const upsMailHTML = "From: Shop <a@shop.example>\r\nSubject: Versendet\r\nContent-Type: text/html; charset=iso-8859-1\r\nContent-Transfer-Encoding: base64\r\n\r\n" +
	"PGEgaHJlZj0iaHR0cHM6Ly93d3cudXBzLmNvbS90cmFjaz90cmFja251bT0xWjk5OUFBMTAxMjM0\r\nNTY3ODQiPlBha2V0IHZlcmZvbGdlbjwvYT4=\r\n"

func TestIMAPReadsNewMailsReadOnly(t *testing.T) {
	f := newFakeIMAP(t)
	f.mails[5] = dhlMail
	f.mails[9] = upsMailHTML
	m := plainIMAP()
	if err := m.Test(context.Background(), f.settings()); err != nil {
		t.Fatal(err)
	}
	mails, st, err := m.Fetch(context.Background(), f.settings(), usecase.MailState{}, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))
	if err != nil || len(mails) != 2 || st.UIDValidity != 7 || st.LastUID != 19 {
		t.Fatalf("%+v %+v %v", mails, st, err)
	}
	if mails[0].Subject != "Ihre Sendung von Bücher Huber" || !strings.Contains(mails[0].Text, "Sendungsnummer: 00340434161094042557") {
		t.Fatalf("decoding: %+v", mails[0])
	}
	if c := domain.ExtractCandidates(mails[1]); len(c) != 1 || c[0].Number != "1Z999AA10123456784" {
		t.Fatalf("base64 HTML link: %q %+v", mails[1].Text, c)
	}
	f.mu.Lock()
	first := strings.Join(f.commands, "\n")
	f.commands = nil
	f.mu.Unlock()
	if !strings.Contains(first, "UID SEARCH SINCE 1-Nov-2026") {
		t.Fatalf("first run looks back: %s", first)
	}
	// Next run: only messages after the last UID.
	f.mails = map[uint32]string{19: upsMailHTML}
	mails, st2, err := m.Fetch(context.Background(), f.settings(), st, time.Now())
	f.mu.Lock()
	defer f.mu.Unlock()
	if err != nil || len(mails) != 0 || st2.LastUID != 19 || !strings.Contains(strings.Join(f.commands, "\n"), "UID SEARCH UID 20:*") {
		t.Fatalf("%v %+v %v %q", mails, st2, err, f.commands)
	}
}

func TestIMAPErrors(t *testing.T) {
	f := newFakeIMAP(t)
	s := f.settings()
	s.Password = "wrong"
	if err := plainIMAP().Test(context.Background(), s); err == nil || !strings.Contains(err.Error(), "login failed") {
		t.Fatal(err)
	}
	s = f.settings()
	s.Folder = "Nope"
	if err := plainIMAP().Test(context.Background(), s); err == nil || !strings.Contains(err.Error(), "Nope") {
		t.Fatal(err)
	}
	// A new UIDVALIDITY starts over.
	f.mails[3] = dhlMail
	mails, st, err := plainIMAP().Fetch(context.Background(), f.settings(), usecase.MailState{UIDValidity: 1, LastUID: 500}, time.Now())
	if err != nil || len(mails) != 1 || st.UIDValidity != 7 {
		t.Fatalf("%v %+v %v", mails, st, err)
	}
}

func TestParseMailSkipsAttachments(t *testing.T) {
	raw := "--x\r\nContent-Type: text/plain\r\n\r\nHallo\r\n--x\r\nContent-Type: application/pdf\r\nContent-Disposition: attachment; filename=r.pdf\r\n\r\nSendungsnummer: 1Z999AA10123456784\r\n--x\r\nContent-Type: text/plain\r\nContent-Disposition: attachment\r\n\r\nSendungsnummer: 1Z999AA10123456784\r\n--x--\r\n"
	m := ParseMail([]byte("From: a@b.c\r\nContent-Type: multipart/mixed; boundary=x\r\n\r\n"), []byte(raw))
	if strings.Contains(m.Text, "1Z999") || !strings.Contains(m.Text, "Hallo") {
		t.Fatalf("%q", m.Text)
	}
	// A body cut off at the fetch limit is read as far as it goes.
	cut := ParseMail([]byte("Content-Type: multipart/mixed; boundary=x\r\n\r\n"), []byte("--x\r\nContent-Type: text/plain\r\n\r\nSendungsnummer: 123"))
	if !strings.Contains(cut.Text, "Sendungsnummer") {
		t.Fatalf("%q", cut.Text)
	}
}
