package gateway

import (
	"sync"
	"time"
)

// maxTrackedClients bounds memory if many distinct clients fail to log in.
const maxTrackedClients = 10000

// AttemptLimiter is an in-memory fixed-window limiter: after MaxFailures
// failed attempts within Window a client is blocked until the window ends.
type AttemptLimiter struct {
	maxFailures int
	window      time.Duration
	mu          sync.Mutex
	entries     map[string]*attempts
}

type attempts struct {
	count       int
	windowStart time.Time
}

// NewAttemptLimiter returns a limiter allowing maxFailures per window.
func NewAttemptLimiter(maxFailures int, window time.Duration) *AttemptLimiter {
	return &AttemptLimiter{maxFailures: maxFailures, window: window, entries: map[string]*attempts{}}
}

// Allow implements usecase.AttemptLimiter.
func (l *AttemptLimiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[key]
	if !ok {
		return true
	}
	if now.Sub(e.windowStart) >= l.window {
		delete(l.entries, key)
		return true
	}
	return e.count < l.maxFailures
}

// RecordFailure implements usecase.AttemptLimiter.
func (l *AttemptLimiter) RecordFailure(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[key]
	if !ok || now.Sub(e.windowStart) >= l.window {
		if len(l.entries) >= maxTrackedClients {
			l.evictExpired(now)
		}
		l.entries[key] = &attempts{count: 1, windowStart: now}
		return
	}
	e.count++
}

// Reset implements usecase.AttemptLimiter.
func (l *AttemptLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

func (l *AttemptLimiter) evictExpired(now time.Time) {
	for k, e := range l.entries {
		if now.Sub(e.windowStart) >= l.window {
			delete(l.entries, k)
		}
	}
}
