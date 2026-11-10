package usecase

import (
	"context"
	"strings"
	"time"

	"trisse-home/internal/modules/auth/domain"
)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

type fakeAccounts struct {
	account *domain.Account
}

func (f *fakeAccounts) Get(context.Context) (domain.Account, error) {
	if f.account == nil {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	return *f.account, nil
}

func (f *fakeAccounts) Create(_ context.Context, a domain.Account) error {
	if f.account != nil {
		return domain.ErrAlreadySetUp
	}
	f.account = &a
	return nil
}

func (f *fakeAccounts) Update(_ context.Context, a domain.Account) error { f.account = &a; return nil }
func (f *fakeAccounts) Delete(context.Context) error                     { f.account = nil; return nil }

type fakeSessions struct {
	byHash map[domain.TokenHash]domain.Session
}

func newFakeSessions() *fakeSessions {
	return &fakeSessions{byHash: map[domain.TokenHash]domain.Session{}}
}

func (f *fakeSessions) Create(_ context.Context, s domain.Session) error {
	f.byHash[s.TokenHash()] = s
	return nil
}

func (f *fakeSessions) Find(_ context.Context, h domain.TokenHash) (domain.Session, error) {
	s, ok := f.byHash[h]
	if !ok {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return s, nil
}

func (f *fakeSessions) Update(_ context.Context, s domain.Session) error {
	f.byHash[s.TokenHash()] = s
	return nil
}

func (f *fakeSessions) Delete(_ context.Context, h domain.TokenHash) error {
	delete(f.byHash, h)
	return nil
}

func (f *fakeSessions) DeleteAll(context.Context) error {
	clear(f.byHash)
	return nil
}

func (f *fakeSessions) DeleteExpired(_ context.Context, now time.Time) (int64, error) {
	var n int64
	for h, s := range f.byHash {
		if s.IsExpired(now) {
			delete(f.byHash, h)
			n++
		}
	}
	return n, nil
}

// fakeHasher "hashes" by prefixing, which keeps tests fast and readable.
type fakeHasher struct{}

func (fakeHasher) Hash(p string) (string, error) { return "hash:" + p, nil }
func (fakeHasher) Verify(p, h string) (bool, error) {
	return strings.TrimPrefix(h, "hash:") == p, nil
}

type seqTokens struct{ n int }

func (s *seqTokens) NewToken() (string, error) {
	s.n++
	return "token-" + string(rune('a'+s.n)), nil
}

type staticSetupToken string

func (s staticSetupToken) Verify(t string) bool { return string(s) == t }

type fakeLimiter struct {
	failures map[string]int
	max      int
}

func newFakeLimiter(max int) *fakeLimiter { return &fakeLimiter{failures: map[string]int{}, max: max} }

func (l *fakeLimiter) Allow(k string, _ time.Time) bool    { return l.failures[k] < l.max }
func (l *fakeLimiter) RecordFailure(k string, _ time.Time) { l.failures[k]++ }
func (l *fakeLimiter) Reset(k string)                      { delete(l.failures, k) }

type countingObserver struct{ calls int }

func (o *countingObserver) AccountCreated(context.Context) { o.calls++ }

type env struct {
	deps     Deps
	accounts *fakeAccounts
	sessions *fakeSessions
	clock    *fakeClock
	limiter  *fakeLimiter
}

func newEnv() *env {
	e := &env{
		accounts: &fakeAccounts{},
		sessions: newFakeSessions(),
		clock:    &fakeClock{now: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)},
		limiter:  newFakeLimiter(3),
	}
	e.deps = Deps{
		Accounts: e.accounts, Sessions: e.sessions, Hasher: fakeHasher{}, Tokens: &seqTokens{},
		Limiter: e.limiter, Clock: e.clock, TTL: SessionTTL(24 * time.Hour),
	}
	return e
}

const goodPassword = "a-very-good-password"

func (e *env) withAccount() *env {
	a, _ := domain.NewAccount("hash:"+goodPassword, e.clock.now)
	e.accounts.account = &a
	return e
}
