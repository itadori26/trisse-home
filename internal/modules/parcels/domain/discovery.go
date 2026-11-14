package domain

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Discovery of parcels in shipping e-mails (#166). The extractors are pure
// functions: a number counts only right after a tracking keyword or inside
// a carrier's tracking link, so order and phone numbers do not match.

// Mail is the part of a shipping e-mail the extractors need. The content is
// never stored.
type Mail struct {
	From    string // display name and address
	Subject string
	Text    string // plain text, or the text and link targets of HTML
}

// Candidate is a tracking number found in a mail.
type Candidate struct {
	Number  string
	Carrier Carrier
	Label   string
}

var (
	// A keyword, then up to 40 characters without a line break, then the number.
	keywordNumber = regexp.MustCompile(`(?i)(?:sendungs-?(?:nummer|nr\.?)|paket-?(?:nummer|nr\.?)|tracking-?(?:nummer|number|code|id)|track(?:ing)? (?:no\.?|#)|sendungsverfolgungsnummer|shipment number|parcel number|ident-?code)[^\n\r0-9A-Z]{0,40}?([0-9A-Z][0-9A-Z ]{6,40}[0-9A-Z])`)
	// Tracking links of the carriers; the number is the first group.
	trackingLinks = []struct {
		carrier Carrier
		re      *regexp.Regexp
	}{
		{CarrierDHL, regexp.MustCompile(`(?i)dhl\.(?:de|com)/[^\s"'<>]*?(?:piececode|idc|tracking-id|trackingnumber|AWB)=([0-9A-Z]{8,35})`)},
		{CarrierHermes, regexp.MustCompile(`(?i)(?:myhermes\.de|hermesworld\.com)/[^\s"'<>]*?(?:#|TrackID=|trackingnumber=)([0-9A-Z]{10,20})`)},
		{CarrierDPD, regexp.MustCompile(`(?i)dpd\.(?:de|com)/[^\s"'<>]*?(?:parcel/|parcelNumber=|query=)([0-9A-Z]{14})`)},
		{CarrierGLS, regexp.MustCompile(`(?i)gls-(?:pakete\.de|group\.(?:eu|com))/[^\s"'<>]*?(?:match=|parcel-?number=)([0-9A-Z]{8,14})`)},
		{CarrierUPS, regexp.MustCompile(`(?i)ups\.com/[^\s"'<>]*?(?:tracknum|InquiryNumber1)=(1Z[0-9A-Z]{16})`)},
	}
	upsNumber   = regexp.MustCompile(`\b1Z[0-9A-Z]{16}\b`)
	carrierName = map[Carrier]*regexp.Regexp{
		CarrierDHL:    regexp.MustCompile(`(?i)\bdhl\b`),
		CarrierHermes: regexp.MustCompile(`(?i)\bhermes\b`),
		CarrierDPD:    regexp.MustCompile(`(?i)\bdpd\b`),
		CarrierGLS:    regexp.MustCompile(`(?i)\bgls\b`),
		CarrierUPS:    regexp.MustCompile(`(?i)\bups\b`),
	}
	fromShop = regexp.MustCompile(`(?i)\b(?:von|from)\s+([^\n\r,.:!?()]{2,40})`)
)

// ExtractCandidates finds the tracking numbers of a shipping e-mail.
func ExtractCandidates(m Mail) []Candidate {
	seen := map[string]bool{}
	var out []Candidate
	add := func(number string, c Carrier) {
		number = strings.ToUpper(strings.ReplaceAll(number, " ", ""))
		if c == "" {
			c = guessCarrier(m, number)
		}
		if c == "" || seen[number] || !ValidNumber(c, number) {
			return
		}
		seen[number] = true
		out = append(out, Candidate{Number: number, Carrier: c, Label: mailLabel(m)})
	}
	text := m.Subject + "\n" + m.Text
	for _, l := range trackingLinks {
		for _, g := range l.re.FindAllStringSubmatch(text, -1) {
			add(g[1], l.carrier)
		}
	}
	for _, g := range keywordNumber.FindAllStringSubmatch(text, -1) {
		add(g[1], "")
	}
	for _, n := range upsNumber.FindAllString(text, -1) {
		add(n, CarrierUPS)
	}
	return out
}

// guessCarrier names the carrier by the number's format and the mail.
func guessCarrier(m Mail, number string) Carrier {
	if strings.HasPrefix(number, "1Z") {
		return CarrierUPS
	}
	if strings.HasPrefix(number, "JJD") || strings.HasPrefix(number, "00340") {
		return CarrierDHL
	}
	// The sender says most, then subject, then text.
	for _, s := range []string{m.From, m.Subject, m.Text} {
		var found Carrier
		for _, c := range Carriers {
			if carrierName[c].MatchString(s) {
				if found != "" {
					found = "ambiguous"
					break
				}
				found = c
			}
		}
		if found != "" && found != "ambiguous" {
			return found
		}
	}
	return ""
}

// ValidNumber checks the format and, where defined, the check digit.
func ValidNumber(c Carrier, n string) bool {
	if !trackingPattern.MatchString(n) {
		return false
	}
	digits := strings.IndexFunc(n, func(r rune) bool { return r < '0' || r > '9' }) < 0
	switch c {
	case CarrierUPS:
		return len(n) == 18 && strings.HasPrefix(n, "1Z") && upsCheck(n)
	case CarrierDHL:
		switch {
		case strings.HasPrefix(n, "JJD"):
			return len(n) >= 13 && len(n) <= 24
		case digits && len(n) == 20 && strings.HasPrefix(n, "00"):
			return gs1Check(n[2:]) // SSCC behind the "00" application identifier
		case digits:
			return len(n) == 12 || len(n) == 14 || len(n) == 20
		}
		return len(n) >= 10 && len(n) <= 14 // international, e.g. CX123456789DE
	case CarrierHermes:
		return digits && (len(n) == 14 || len(n) == 16) || (len(n) == 20 && strings.HasPrefix(n, "H"))
	case CarrierDPD:
		return digits && len(n) == 14 || len(n) == 14 && unicode.IsLetter(rune(n[13]))
	case CarrierGLS:
		return digits && (len(n) == 11 || len(n) == 12) || (len(n) == 8 && !digits)
	}
	return false
}

// gs1Check validates the GS1 mod-10 check digit (last digit).
func gs1Check(s string) bool {
	sum := 0
	for i := len(s) - 2; i >= 0; i-- {
		d := int(s[i] - '0')
		if (len(s)-2-i)%2 == 0 {
			d *= 3
		}
		sum += d
	}
	return (10-sum%10)%10 == int(s[len(s)-1]-'0')
}

// upsCheck validates the check digit of a 1Z number.
func upsCheck(s string) bool {
	body := s[2:17]
	sum := 0
	for i, r := range body {
		var v int
		if r >= '0' && r <= '9' {
			v = int(r - '0')
		} else {
			v = int(r-'A'+2) % 10 // letters map to digits as UPS defines
		}
		if i%2 == 1 {
			v *= 2
		}
		sum += v
	}
	check := (10 - sum%10) % 10
	return strconv.Itoa(check) == s[17:]
}

// mailLabel names the parcel: the shop that sent the mail, or the shop a
// carrier mail names ("Ihre Sendung von Zalando"), or the subject.
func mailLabel(m Mail) string {
	name := m.From
	if i := strings.Index(name, "<"); i >= 0 {
		name = name[:i]
	}
	name = strings.Trim(strings.TrimSpace(name), `"'`)
	isCarrier := false
	for _, c := range Carriers {
		if carrierName[c].MatchString(m.From) {
			isCarrier = true
		}
	}
	switch {
	case !isCarrier && name != "" && !strings.Contains(name, "@"):
		return clip(name)
	case fromShop.MatchString(m.Subject):
		return clip(strings.TrimSpace(fromShop.FindStringSubmatch(m.Subject)[1]))
	}
	return clip(strings.TrimSpace(m.Subject))
}

func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) > maxLabelLength {
		s = string([]rune(s)[:maxLabelLength-1]) + "…"
	}
	return s
}

// MailSettings configure the read-only mailbox the parcels are found in.
type MailSettings struct {
	Enabled  bool
	Host     string
	Port     int
	Username string
	Password string
	Folder   string
	Interval time.Duration
}

// Limits of the mail discovery.
const (
	DefaultMailInterval = 15 * time.Minute
	minMailInterval     = 5 * time.Minute
	maxMailInterval     = 24 * time.Hour
	DefaultMailPort     = 993
	// FirstRunDays is how far back the first run looks.
	FirstRunDays = 14
)

var hostName = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.-]{0,251}[A-Za-z0-9])?$`)

// NewMailSettings validates the mailbox settings.
func NewMailSettings(enabled bool, host string, port int, user, password, folder string, interval time.Duration) (MailSettings, error) {
	s := MailSettings{Enabled: enabled, Host: strings.ToLower(strings.TrimSpace(host)), Port: port, Username: strings.TrimSpace(user),
		Password: password, Folder: strings.TrimSpace(folder), Interval: interval}
	if s.Port == 0 {
		s.Port = DefaultMailPort
	}
	if s.Interval == 0 {
		s.Interval = DefaultMailInterval
	}
	if s.Folder == "" {
		s.Folder = "INBOX"
	}
	switch {
	case !enabled && s.Host == "":
		return s, nil
	case !hostName.MatchString(s.Host):
		return MailSettings{}, kernel.NewValidationError("host", "Enter the IMAP server name, e.g. imap.example.com.")
	case s.Port < 1 || s.Port > 65535:
		return MailSettings{}, kernel.NewValidationError("port", "The port must be between 1 and 65535 (IMAP over TLS: 993).")
	case s.Username == "" || s.Password == "":
		return MailSettings{}, kernel.NewValidationError("username", "User name and app password are required.")
	case !imapSafe(s.Username) || !imapSafe(s.Password) || !imapSafe(s.Folder) || len(s.Folder) > 200:
		return MailSettings{}, kernel.NewValidationError("folder", "User name, password and folder must not contain line breaks.")
	case s.Interval < minMailInterval || s.Interval > maxMailInterval:
		return MailSettings{}, kernel.NewValidationError("intervalMinutes", "Check the mailbox every 5 minutes to 24 hours.")
	}
	return s, nil
}

// imapSafe reports whether s can be sent as an IMAP quoted string (quotes
// and backslashes are escaped by the client; line breaks cannot be).
func imapSafe(s string) bool {
	return !strings.ContainsAny(s, "\r\n\x00")
}
