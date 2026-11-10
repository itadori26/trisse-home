package scheduler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerRunsImmediatelyAndRepeats(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var runs atomic.Int32
	s := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	s.Add(Task{Name: "count", Interval: 10 * time.Millisecond, Run: func(context.Context) error {
		runs.Add(1)
		return nil
	}})
	s.Add(Task{Name: "fails", Interval: 10 * time.Millisecond, Run: func(context.Context) error {
		return errors.New("boom")
	}})
	s.Add(Task{Name: "panics", Interval: 10 * time.Millisecond, Run: func(context.Context) error {
		panic("boom")
	}})
	s.Start(ctx)
	deadline := time.After(2 * time.Second)
	for runs.Load() < 3 {
		select {
		case <-deadline:
			t.Fatalf("only %d runs", runs.Load())
		case <-time.After(5 * time.Millisecond):
		}
	}
	cancel()
	s.Wait()
}
