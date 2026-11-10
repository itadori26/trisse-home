package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNewNote(t *testing.T) {
	now := time.Now()
	n, err := NewNote("1", "", "\n# Shopping list\n- milk", false, now)
	if err != nil || n.Title() != "Shopping list" {
		t.Fatalf("title not derived: %q %v", n.Title(), err)
	}
	if _, err := NewNote("1", " ", "  ", false, now); err == nil {
		t.Fatal("expected error for empty note")
	}
	if _, err := NewNote("1", "x", strings.Repeat("a", MaxBodyBytes+1), false, now); err == nil {
		t.Fatal("expected size error")
	}
	long, _ := NewNote("1", strings.Repeat("ä", 300), "", false, now)
	if len([]rune(long.Title())) != MaxTitleLength {
		t.Fatal("long title not truncated")
	}
}
