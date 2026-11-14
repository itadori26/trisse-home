package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"net/netip"
	"strings"
	"unicode"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Summaries with a local Ollama (#75): the text never leaves the home
// network, so only Ollama addresses in the LAN are accepted.

// Limits of the input.
const (
	DefaultMaxInputChars = 12000
	MinMaxInputChars     = 1000
	MaxMaxInputChars     = 50000
	maxModelLength       = 100
	// MaxSummaryChars bounds what is stored and shown.
	MaxSummaryChars = 4000
)

// Errors.
var (
	ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "Summaries are not set up; add your Ollama server in the settings.")
	ErrNothingToSum  = kernel.NewValidationError("text", "There is no text to summarise.")
)

// Settings connect a local Ollama.
type Settings struct {
	URL           string
	Model         string
	MaxInputChars int
}

// NewSettings validates the settings.
func NewSettings(rawURL, model string, maxInput int) (Settings, error) {
	u, err := kernel.ParseTarget("url", rawURL, kernel.WithTargetPath())
	if err != nil {
		return Settings{}, err
	}
	if !LocalHost(u.Hostname()) {
		return Settings{}, kernel.NewValidationError("url",
			"Only an Ollama in your home network is allowed, e.g. http://192.168.1.10:11434 or http://ollama.lan:11434.")
	}
	model = strings.TrimSpace(model)
	if model == "" || len(model) > maxModelLength || strings.ContainsFunc(model, unicode.IsSpace) {
		return Settings{}, kernel.NewValidationError("model", "Choose a model of your Ollama server.")
	}
	if maxInput == 0 {
		maxInput = DefaultMaxInputChars
	}
	if maxInput < MinMaxInputChars || maxInput > MaxMaxInputChars {
		return Settings{}, kernel.NewValidationError("maxInputChars", "The input limit must be between 1,000 and 50,000 characters.")
	}
	return Settings{URL: u.String(), Model: model, MaxInputChars: maxInput}, nil
}

// localSuffixes are names that only resolve in a home network.
var localSuffixes = []string{".local", ".lan", ".home", ".home.arpa", ".internal", ".localdomain", ".fritz.box"}

// LocalHost reports whether host is in the home network: a loopback or
// private address, or a name without a public domain.
func LocalHost(host string) bool {
	host = strings.ToLower(strings.Trim(host, "[]"))
	if ip, err := netip.ParseAddr(host); err == nil {
		ip = ip.Unmap()
		return ip.IsLoopback() || ip.IsPrivate()
	}
	if host == "localhost" || !strings.Contains(host, ".") {
		return true // single-label names such as "ollama" (Docker, mDNS)
	}
	for _, s := range localSuffixes {
		if strings.HasSuffix(host, s) {
			return true
		}
	}
	return false
}

// Language of the summary.
type Language string

// Supported languages; anything else falls back to English.
const (
	LanguageEN Language = "en"
	LanguageDE Language = "de"
)

// ParseLanguage maps the UI language.
func ParseLanguage(raw string) Language {
	if strings.HasPrefix(strings.ToLower(raw), "de") {
		return LanguageDE
	}
	return LanguageEN
}

// Input is the text to summarise, normalised and cut to the limit.
type Input struct {
	Text      string
	Truncated bool
}

// NewInput collapses whitespace and cuts the text at a word boundary.
func NewInput(text string, maxChars int) (Input, error) {
	text = strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(text) < 20 {
		return Input{}, ErrNothingToSum
	}
	if utf8.RuneCountInString(text) <= maxChars {
		return Input{Text: text}, nil
	}
	cut := string([]rune(text)[:maxChars])
	if i := strings.LastIndexByte(cut, ' '); i > maxChars/2 {
		cut = cut[:i]
	}
	return Input{Text: cut, Truncated: true}, nil
}

// CacheKey identifies a summary: the same text, model and language give
// the same summary, a changed note a new one.
func CacheKey(model string, lang Language, text string) string {
	sum := sha256.Sum256([]byte(model + "\x00" + string(lang) + "\x00" + text))
	return hex.EncodeToString(sum[:])
}

// Prompt asks for a short summary in the language.
func Prompt(lang Language, text string) string {
	if lang == LanguageDE {
		return "Fasse den folgenden Text in 3 bis 5 kurzen Stichpunkten auf Deutsch zusammen. " +
			"Antworte nur mit den Stichpunkten, ohne Einleitung.\n\n" + text
	}
	return "Summarise the following text in 3 to 5 short bullet points in English. " +
		"Answer only with the bullet points, without an introduction.\n\n" + text
}

// CleanSummary trims the model's answer and bounds its length.
func CleanSummary(s string) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > MaxSummaryChars {
		s = string([]rune(s)[:MaxSummaryChars-1]) + "…"
	}
	return s
}

// Asking the model from the smart search (#162): only the question and a
// compact status context that the frontend builds from what the user can
// see anyway.

// Limits of a question.
const (
	MaxQuestionChars = 300
	MaxContextChars  = 6000
)

// Question is a free question with its context.
type Question struct {
	Text    string
	Context string
}

// NewQuestion validates a question; the context is cut to its limit.
func NewQuestion(text, context string) (Question, error) {
	text = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), "?"))
	if utf8.RuneCountInString(text) < 3 {
		return Question{}, kernel.NewValidationError("question", "Ask a question after the question mark.")
	}
	if utf8.RuneCountInString(text) > MaxQuestionChars {
		return Question{}, kernel.NewValidationError("question", "The question must be at most 300 characters long.")
	}
	context = strings.TrimSpace(context)
	if utf8.RuneCountInString(context) > MaxContextChars {
		context = string([]rune(context)[:MaxContextChars])
	}
	return Question{Text: text, Context: context}, nil
}

// AskPrompt asks for a short answer grounded in the context.
func AskPrompt(lang Language, q Question) string {
	if lang == LanguageDE {
		return "Du bist der Assistent eines Heimserver-Dashboards. Beantworte die Frage kurz (höchstens 5 Sätze) auf Deutsch " +
			"und nur anhand des Kontexts. Wenn der Kontext die Antwort nicht enthält, sag das ehrlich.\n\nKontext:\n" +
			q.Context + "\n\nFrage: " + q.Text
	}
	return "You are the assistant of a home server dashboard. Answer the question briefly (at most 5 sentences) in English, " +
		"using only the context. If the context does not contain the answer, say so honestly.\n\nContext:\n" +
		q.Context + "\n\nQuestion: " + q.Text
}
