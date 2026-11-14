package domain

import (
	"cmp"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits.
const (
	MaxRules       = 20
	maxMatchLength = 40
	maxLabelLength = 40
	maxIconLength  = 8
	// UpcomingPerType is how many dates are listed per type.
	UpcomingPerType = 3
)

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// ErrNotConfigured is returned while no calendar (URL or file) is configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "No waste calendar is configured yet.")

// Rule maps summaries containing Match (case-insensitive) to a label and
// style; empty fields keep the automatic values.
type Rule struct {
	Match string
	Label string
	Color string
	Icon  string
}

// Settings are the source of the dates (ICS URL or uploaded file, #153)
// and the custom rules. Rules are kept when the source changes.
type Settings struct {
	Source Source
	Rules  []Rule
}

// NewSettings validates settings with an ICS subscription URL.
func NewSettings(rawURL string, rules []Rule) (Settings, error) {
	src, err := NewURLSource(rawURL)
	if err != nil {
		return Settings{}, err
	}
	clean, err := NewRules(rules)
	if err != nil {
		return Settings{}, err
	}
	return Settings{Source: src, Rules: clean}, nil
}

func normaliseURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if strings.HasPrefix(strings.ToLower(rawURL), "webcal://") {
		rawURL = "https://" + rawURL[len("webcal://"):]
	}
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", kernel.NewValidationError("url", "The ICS URL must start with https://, http:// or webcal://.")
	}
	return u.String(), nil
}

// NewRules validates and trims the custom rules.
func NewRules(rules []Rule) ([]Rule, error) {
	if len(rules) > MaxRules {
		return nil, kernel.NewValidationError("rules", "At most 20 rules are supported.")
	}
	clean := make([]Rule, 0, len(rules))
	for _, r := range rules {
		r = Rule{Match: strings.TrimSpace(r.Match), Label: strings.TrimSpace(r.Label), Color: strings.TrimSpace(r.Color), Icon: strings.TrimSpace(r.Icon)}
		switch {
		case r.Match == "" || utf8.RuneCountInString(r.Match) > maxMatchLength:
			return nil, kernel.NewValidationError("rules", "Every rule needs a match text of at most 40 characters.")
		case utf8.RuneCountInString(r.Label) > maxLabelLength:
			return nil, kernel.NewValidationError("rules", "Labels must be at most 40 characters long.")
		case r.Color != "" && !colorPattern.MatchString(r.Color):
			return nil, kernel.NewValidationError("rules", "Colors must look like #1a2b3c.")
		case utf8.RuneCountInString(r.Icon) > maxIconLength:
			return nil, kernel.NewValidationError("rules", "Icons must be at most 8 characters long.")
		}
		clean = append(clean, r)
	}
	return clean, nil
}

// Style is how a collection type is shown.
type Style struct {
	Label string
	Color string
	Icon  string
}

type builtin struct {
	keywords []string
	color    string
	icon     string
}

// builtins recognise common German and English summaries. Order matters:
// more specific keywords come first.
var builtins = []builtin{
	{[]string{"grünschnitt", "gruenschnitt", "garden", "baum"}, "#16a34a", "🌳"},
	{[]string{"schadstoff", "hazard"}, "#dc2626", "☣️"},
	{[]string{"sperr", "bulky"}, "#a855f7", "🛋️"},
	{[]string{"bio", "organic", "kompost"}, "#8b5a2b", "🍂"},
	{[]string{"papier", "pappe", "paper", "blau"}, "#2563eb", "📦"},
	{[]string{"gelb", "wertstoff", "verpackung", "plastic", "recycl"}, "#eab308", "♻️"},
	{[]string{"glas", "glass"}, "#10b981", "🍾"},
	{[]string{"rest", "residual", "grau", "schwarz", "general"}, "#6b7280", "🗑️"},
}

// Classify returns the style of a summary: the first matching custom rule
// wins, then the built-in keywords; the label defaults to the summary.
func Classify(summary string, rules []Rule) Style {
	summary = strings.TrimSpace(summary)
	lower := strings.ToLower(summary)
	s := Style{Label: summary, Color: "#64748b", Icon: "🗓️"}
	for _, b := range builtins {
		if slices.ContainsFunc(b.keywords, func(k string) bool { return strings.Contains(lower, k) }) {
			s.Color, s.Icon = b.color, b.icon
			break
		}
	}
	for _, r := range rules {
		if strings.Contains(lower, strings.ToLower(r.Match)) {
			if r.Label != "" {
				s.Label = r.Label
			}
			if r.Color != "" {
				s.Color = r.Color
			}
			if r.Icon != "" {
				s.Icon = r.Icon
			}
			break
		}
	}
	return s
}

// Collection is one pick-up date (a calendar date at midnight UTC).
type Collection struct {
	Date    time.Time
	Summary string
}

// Date returns the calendar date of t (in t's location) as midnight UTC.
func Date(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Upcoming groups the collections from today on by type.
type Upcoming struct {
	Style     Style
	Dates     []time.Time
	DaysUntil int
}

// NextCollections returns, per type (label), the next dates from today on,
// ordered by the next date. Duplicates on the same day are merged.
func NextCollections(cs []Collection, rules []Rule, today time.Time) []Upcoming {
	today = Date(today)
	byLabel := map[string]*Upcoming{}
	slices.SortFunc(cs, func(a, b Collection) int { return a.Date.Compare(b.Date) })
	for _, c := range cs {
		if c.Date.Before(today) {
			continue
		}
		st := Classify(c.Summary, rules)
		key := strings.ToLower(st.Label)
		u, ok := byLabel[key]
		if !ok {
			u = &Upcoming{Style: st, DaysUntil: int(c.Date.Sub(today).Hours() / 24)}
			byLabel[key] = u
		}
		if len(u.Dates) < UpcomingPerType && !slices.ContainsFunc(u.Dates, c.Date.Equal) {
			u.Dates = append(u.Dates, c.Date)
		}
	}
	out := make([]Upcoming, 0, len(byLabel))
	for _, u := range byLabel {
		out = append(out, *u)
	}
	slices.SortFunc(out, func(a, b Upcoming) int {
		return cmp.Or(a.Dates[0].Compare(b.Dates[0]), cmp.Compare(a.Style.Label, b.Style.Label))
	})
	return out
}
