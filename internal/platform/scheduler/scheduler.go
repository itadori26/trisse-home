package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Task is a named function executed every Interval. It runs once right after
// Start so caches are warm before the first request arrives.
type Task struct {
	Name     string
	Interval time.Duration
	Run      func(ctx context.Context) error
}

// Scheduler executes tasks until its context is cancelled.
type Scheduler struct {
	log   *slog.Logger
	tasks []Task
	wg    sync.WaitGroup
}

// New returns a Scheduler that logs to log.
func New(log *slog.Logger) *Scheduler { return &Scheduler{log: log} }

// Add registers a task. It must be called before Start.
func (s *Scheduler) Add(t Task) { s.tasks = append(s.tasks, t) }

// Start launches one goroutine per task and returns immediately.
func (s *Scheduler) Start(ctx context.Context) {
	for _, t := range s.tasks {
		s.wg.Add(1)
		go s.loop(ctx, t)
	}
	s.log.Info("scheduler started", "tasks", len(s.tasks))
}

// Wait blocks until all task goroutines have stopped.
func (s *Scheduler) Wait() { s.wg.Wait() }

func (s *Scheduler) loop(ctx context.Context, t Task) {
	defer s.wg.Done()
	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()
	for {
		s.runOnce(ctx, t)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context, t Task) {
	defer func() {
		if rec := recover(); rec != nil {
			s.log.Error("job panicked", "job", t.Name, "panic", rec)
		}
	}()
	// A job may never take longer than its interval, otherwise runs pile up.
	runCtx, cancel := context.WithTimeout(ctx, t.Interval)
	defer cancel()
	start := time.Now()
	if err := t.Run(runCtx); err != nil && ctx.Err() == nil {
		s.log.Warn("job failed", "job", t.Name, "err", err)
		return
	}
	s.log.Debug("job finished", "job", t.Name, "duration_ms", time.Since(start).Milliseconds())
}
