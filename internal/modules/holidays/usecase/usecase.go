package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/holidays/domain"
	"trisse-home/internal/shared/kernel"
)

// Overview is returned by Upcoming.
type Overview struct {
	State   domain.State
	Entries []domain.Entry
	// School is true when an ICS link for school holidays is configured.
	School bool
	// SchoolUpdatedAt is when the school holidays were last loaded.
	SchoolUpdatedAt time.Time
	// SchoolError is set when the last refresh failed (the cached dates are still shown).
	SchoolError string
}

// Holidays is the input port of the module.
type Holidays interface {
	Settings(ctx context.Context) (domain.Settings, error)
	// SaveSettings stores the ICS link; a new link is loaded first so typos
	// are reported immediately. An empty link removes the school holidays.
	SaveSettings(ctx context.Context, schoolURL string) (domain.Settings, error)
	Upcoming(ctx context.Context, state domain.State, n int) (Overview, error)
	// Refresh reloads the school holidays when the cache is a week old (background job).
	Refresh(ctx context.Context) error
}

// SettingsRepository stores the settings (output port).
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// SchoolSource loads school holidays from an ICS link (output port).
type SchoolSource interface {
	Fetch(ctx context.Context, url string, from, to time.Time) ([]domain.SchoolHoliday, error)
}

// SchoolCache keeps the loaded school holidays, stored per year (output port).
type SchoolCache interface {
	// Load returns the cached periods, the link they came from and when they
	// were loaded; an empty cache returns a zero time.
	Load(ctx context.Context) (periods []domain.SchoolHoliday, url string, at time.Time, err error)
	Store(ctx context.Context, url string, periods []domain.SchoolHoliday, at time.Time) error
	Clear(ctx context.Context) error
}

// maxAge is how long loaded school holidays are used before they are
// loaded again; they change rarely (once a year in practice).
const maxAge = 7 * 24 * time.Hour

// window is the period that is loaded and cached: last year to the year after next.
func window(now time.Time) (time.Time, time.Time) {
	y := now.Year()
	return time.Date(y-1, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(y+2, 1, 1, 0, 0, 0, 0, time.UTC)
}

type holidays struct {
	repo   SettingsRepository
	source SchoolSource
	cache  SchoolCache
	clock  kernel.Clock

	mu        sync.Mutex
	lastError string
}

// New returns the use cases.
func New(repo SettingsRepository, source SchoolSource, cache SchoolCache, clock kernel.Clock) Holidays {
	return &holidays{repo: repo, source: source, cache: cache, clock: clock}
}

func (uc *holidays) Settings(ctx context.Context) (domain.Settings, error) { return uc.repo.Load(ctx) }

func (uc *holidays) SaveSettings(ctx context.Context, schoolURL string) (domain.Settings, error) {
	s, err := domain.NewSettings(schoolURL)
	if err != nil {
		return domain.Settings{}, err
	}
	if s.SchoolURL == "" {
		if err := uc.cache.Clear(ctx); err != nil {
			return domain.Settings{}, fmt.Errorf("clear cache: %w", err)
		}
	} else if err := uc.load(ctx, s.SchoolURL); err != nil {
		return domain.Settings{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not load the school holidays: {detail}", "detail", err.Error())
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return domain.Settings{}, fmt.Errorf("save settings: %w", err)
	}
	return s, nil
}

func (uc *holidays) Upcoming(ctx context.Context, state domain.State, n int) (Overview, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return Overview{}, fmt.Errorf("load settings: %w", err)
	}
	o := Overview{State: state, School: s.SchoolURL != ""}
	var school []domain.SchoolHoliday
	if o.School {
		periods, url, at, err := uc.cache.Load(ctx)
		if err != nil {
			return Overview{}, fmt.Errorf("load cache: %w", err)
		}
		if url != s.SchoolURL || at.IsZero() {
			// Nothing cached for this link yet (e.g. after a restore): load once now.
			_ = uc.Refresh(ctx)
			periods, _, at, _ = uc.cache.Load(ctx)
		}
		school, o.SchoolUpdatedAt = periods, at
		uc.mu.Lock()
		o.SchoolError = uc.lastError
		uc.mu.Unlock()
	}
	o.Entries = domain.Upcoming(uc.clock.Now(), state, school, n)
	return o, nil
}

func (uc *holidays) Refresh(ctx context.Context) error {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	if s.SchoolURL == "" {
		return nil
	}
	_, url, at, err := uc.cache.Load(ctx)
	if err != nil {
		return fmt.Errorf("load cache: %w", err)
	}
	if url == s.SchoolURL && uc.clock.Now().Sub(at) < maxAge {
		return nil
	}
	if err := uc.load(ctx, s.SchoolURL); err != nil {
		uc.mu.Lock()
		uc.lastError = "The school holidays could not be loaded; showing the last known dates."
		uc.mu.Unlock()
		return fmt.Errorf("load school holidays: %w", err)
	}
	return nil
}

// load fetches and caches the school holidays of url.
func (uc *holidays) load(ctx context.Context, url string) error {
	now := uc.clock.Now()
	from, to := window(now)
	periods, err := uc.source.Fetch(ctx, url, from, to)
	if err != nil {
		return err
	}
	if len(periods) == 0 {
		return errors.New("the calendar contains no holidays in this and the next year")
	}
	if err := uc.cache.Store(ctx, url, periods, now); err != nil {
		return fmt.Errorf("store cache: %w", err)
	}
	uc.mu.Lock()
	uc.lastError = ""
	uc.mu.Unlock()
	return nil
}
