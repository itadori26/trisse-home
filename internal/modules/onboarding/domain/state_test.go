package domain

import (
	"testing"
	"time"
)

func TestState(t *testing.T) {
	var s State
	if s.Completed() {
		t.Fatal("zero state must be pending")
	}
	first := time.Unix(100, 0)
	s.Complete(first)
	s.Complete(time.Unix(200, 0))
	if !s.Completed() || !s.CompletedAt().Equal(first) {
		t.Fatalf("unexpected %v", s.CompletedAt())
	}
	s.Restart()
	if s.Completed() {
		t.Fatal("restart must make the tour pending again")
	}
}
