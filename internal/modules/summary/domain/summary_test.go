package domain

import (
	"errors"
	"strings"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestSettings(t *testing.T) {
	s, err := NewSettings("http://192.168.1.10:11434/", "llama3.2:3b", 0)
	if err != nil || s.URL != "http://192.168.1.10:11434" || s.MaxInputChars != DefaultMaxInputChars {
		t.Fatalf("%+v %v", s, err)
	}
	for _, ok := range []string{"http://localhost:11434", "http://ollama:11434", "http://ollama.lan:11434", "http://nas.fritz.box:11434", "http://[::1]:11434", "http://10.0.0.5"} {
		if _, err := NewSettings(ok, "m", 0); err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	for _, bad := range []string{"https://api.example.com", "http://8.8.8.8:11434", "http://169.254.169.254", "ftp://ollama.lan"} {
		if _, err := NewSettings(bad, "m", 0); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s accepted", bad)
		}
	}
	if _, err := NewSettings("http://ollama:11434", "has space", 0); err == nil {
		t.Error("model with space")
	}
	if _, err := NewSettings("http://ollama:11434", "m", 100); err == nil {
		t.Error("tiny input limit")
	}
}

func TestInput(t *testing.T) {
	if _, err := NewInput("  zu kurz ", 1000); !errors.Is(err, ErrNothingToSum) {
		t.Fatal(err)
	}
	in, err := NewInput(strings.Repeat("Wort ", 400), 1000)
	if err != nil || !in.Truncated || len(in.Text) > 1000 || strings.HasSuffix(in.Text, " ") || !strings.HasSuffix(in.Text, "Wort") {
		t.Fatalf("%v %q", err, in.Text[len(in.Text)-10:])
	}
	if a, b := CacheKey("m", LanguageDE, "x"), CacheKey("m", LanguageEN, "x"); a == b || len(a) != 64 {
		t.Fatal("language is part of the key")
	}
	if ParseLanguage("de-DE") != LanguageDE || ParseLanguage("fr") != LanguageEN {
		t.Fatal("language")
	}
	if !strings.Contains(Prompt(LanguageDE, "T"), "Deutsch") || !strings.HasSuffix(Prompt(LanguageEN, "T"), "\n\nT") {
		t.Fatal("prompt")
	}
	if got := CleanSummary("  " + strings.Repeat("a", MaxSummaryChars+10)); len([]rune(got)) != MaxSummaryChars {
		t.Fatal(len(got))
	}
}

func TestQuestion(t *testing.T) {
	q, err := NewQuestion(" ? Warum war jellyfin offline? ", strings.Repeat("x", MaxContextChars+10))
	if err != nil || q.Text != "Warum war jellyfin offline?" || len(q.Context) != MaxContextChars {
		t.Fatalf("%+v %v", q.Text, err)
	}
	for _, bad := range []string{"?", "? a", "?" + strings.Repeat("a", MaxQuestionChars+1)} {
		if _, err := NewQuestion(bad, ""); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%q accepted", bad)
		}
	}
	p := AskPrompt(LanguageDE, Question{Text: "F?", Context: "service jellyfin: offline"})
	if !strings.Contains(p, "Kontext:\nservice jellyfin: offline") || !strings.HasSuffix(p, "Frage: F?") {
		t.Fatal(p)
	}
}
