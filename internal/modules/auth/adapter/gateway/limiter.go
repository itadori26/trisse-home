package gateway

import (
	"sync"
	"time"

	"trisse-home/internal/modules/auth/domain"
)

// maxTrackedClients bounds memory if many distinct clients fail to log in.
const maxTrackedClients = 10000

// AttemptLimiter implements usecase.AttemptLimiter in memory with the
// domain.LoginThrottle policy (lockout with exponential backoff).
type AttemptLimiter struct {
	policy  domain.LoginThrottle
	mu      sync.Mutex
	entries map[string]*attempts
}

type attempts struct {
	count       int
	windowStart time.Time
	lockedUntil time.Time
	lockouts    int
}

// NewAttemptLimiter returns a limiter with the given policy.
func NewAttemptLimiter(policy domain.LoginThrottle) *AttemptLimiter {
	return &AttemptLimiter{policy: policy, entries: map[string]*attempts{}}
}

// Allow implements usecase.AttemptLimiter.
func (l *AttemptLimiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[key]
	return !ok || !now.Before(e.lockedUntil)
}

// RecordFailure implements usecase.AttemptLimiter and reports whether this
// failure locked the client out.
func (l *AttemptLimiter) RecordFailure(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[key]
	if !ok {
		if len(l.entries) >= maxTrackedClients {
			l.evictIdle(now)
		}
		e = &attempts{windowStart: now}
		l.entries[key] = e
	}
	if now.Sub(e.windowStart) >= l.policy.Window {
		e.count, e.windowStart = 0, now
	}
	e.count++
	if e.count < l.policy.MaxFailures {
		return false
	}
	e.lockedUntil = now.Add(l.policy.Lockout(e.lockouts))
	e.lockouts++
	e.count, e.windowStart = 0, now
	return true
}

// Reset implements usecase.AttemptLimiter (after a successful login).
func (l *AttemptLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

// evictIdle drops clients that are neither locked nor inside a window.
func (l *AttemptLimiter) evictIdle(now time.Time) {
	for k, e := range l.entries {
		if !now.Before(e.lockedUntil) && now.Sub(e.windowStart) >= l.policy.Window {
			delete(l.entries, k)
		}
	}
}
