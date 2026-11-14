package domain

import (
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestState(t *testing.T) {
	if _, err := ParseMode("desktop"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
	for _, c := range []struct {
		s    State
		want Phase
	}{
		{State{}, PhaseIdle},
		{State{Requested: ModeKiosk}, PhasePending},
		{State{Requested: ModeKiosk, Progress: Progress{Mode: ModeHeadless, Phase: PhaseActive}}, PhasePending},
		{State{Requested: ModeKiosk, Progress: Progress{Mode: ModeKiosk, Phase: PhaseInstalling}}, PhaseInstalling},
		{State{Requested: ModeKiosk, Progress: Progress{Mode: ModeKiosk, Phase: PhaseFailed, Error: "no internet"}}, PhaseFailed},
	} {
		if got := c.s.Phase(); got != c.want {
			t.Errorf("%+v: %s, want %s", c.s, got, c.want)
		}
	}
	busy := State{Requested: ModeKiosk, Progress: Progress{Mode: ModeKiosk, Phase: PhaseInstalling}}
	if !errors.Is(busy.CanChange(), ErrBusy) {
		t.Fatal("busy")
	}
}
