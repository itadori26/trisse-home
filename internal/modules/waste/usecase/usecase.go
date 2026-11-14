package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
	"unicode/utf8"

	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/shared/kernel"
)

// Overview is returned by Upcoming.
type Overview struct {
	Types     []domain.Upcoming
	UpdatedAt time.Time
	Error     string
	// EndsOn is the last date of an uploaded calendar that ends within
	// domain.ExpiryWarning (#153); zero otherwise.
	EndsOn time.Time
}

// Waste is the input port of the module.
type Waste interface {
	Settings(ctx context.Context) (domain.Settings, bool, error)
	// SaveSettings stores the rules and, if rawURL is not empty, switches to
	// that ICS subscription. An empty rawURL keeps the current source.
	SaveSettings(ctx context.Context, rawURL string, rules []domain.Rule) (domain.Settings, error)
	// UploadCalendar validates an ICS file and makes it the source (#153).
	UploadCalendar(ctx context.Context, name string, content []byte) (domain.Settings, error)
	// RemoveCalendar removes an uploaded file; the rules stay.
	RemoveCalendar(ctx context.Context) (domain.Settings, error)
	Upcoming(ctx context.Context) (Overview, error)
	Refresh(ctx context.Context) error
}

// SettingsRepository persists the settings and the uploaded file (output port).
type SettingsRepository interface {
	// Load returns the settings; without a stored row ErrNotConfigured.
	// A row whose source was removed has Source.Kind SourceNone.
	Load(ctx context.Context) (domain.Settings, error)
	// Save stores settings with a URL source or without a source; a stored
	// file is deleted in the same transaction.
	Save(ctx context.Context, s domain.Settings) error
	// SaveWithFile stores settings with a file source together with the file
	// content, atomically (#153); nil content keeps the stored file.
	SaveWithFile(ctx context.Context, s domain.Settings, content []byte) error
}

// CollectionSource loads the collection dates of a source (output port):
// the ICS URL or the stored file.
type CollectionSource interface {
	Collections(ctx context.Context, src domain.Source, from, to time.Time) ([]domain.Collection, error)
}

// CalendarParser reads the collection dates out of ICS content (output port),
// so an upload is validated without the network gateway.
type CalendarParser interface {
	Parse(content []byte, from, to time.Time) ([]domain.Collection, error)
}

// horizon is how far ahead dates are loaded.
const horizon = 366 * 24 * time.Hour

// fileWindow is the period an uploaded file is summarised over (event count,
// first and last date); recurring rules without an end are cut there.
const fileWindow = 2 * 366 * 24 * time.Hour

type waste struct {
	repo   SettingsRepository
	source CollectionSource
	parser CalendarParser
	clock  kernel.Clock

	mu        sync.RWMutex
	loaded    bool
	cs        []domain.Collection
	updatedAt time.Time
	lastError string
}

// NewWaste returns the waste use cases.
func NewWaste(repo SettingsRepository, source CollectionSource, parser CalendarParser, clock kernel.Clock) Waste {
	return &waste{repo: repo, source: source, parser: parser, clock: clock}
}

// Settings implements Waste; configured is false without a source.
func (uc *waste) Settings(ctx context.Context) (domain.Settings, bool, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return domain.Settings{Rules: []domain.Rule{}}, false, nil
	}
	if err != nil {
		return domain.Settings{}, false, fmt.Errorf("load settings: %w", err)
	}
	return s, s.Source.Configured(), nil
}

// SaveSettings implements Waste; a new subscription is loaded first so typos
// are reported immediately.
func (uc *waste) SaveSettings(ctx context.Context, rawURL string, rules []domain.Rule) (domain.Settings, error) {
	clean, err := domain.NewRules(rules)
	if err != nil {
		return domain.Settings{}, err
	}
	if rawURL == "" {
		current, err := uc.repo.Load(ctx)
		if err != nil && !errors.Is(err, domain.ErrNotConfigured) {
			return domain.Settings{}, fmt.Errorf("load settings: %w", err)
		}
		if !current.Source.Configured() {
			return domain.Settings{}, kernel.NewValidationError("url", "Enter an ICS URL or upload an ICS file.")
		}
		current.Rules = clean
		save := uc.repo.Save
		if current.Source.Kind == domain.SourceFile {
			// Only the rules change; the stored file stays.
			save = uc.saveRulesKeepingFile
		}
		if err := save(ctx, current); err != nil {
			return domain.Settings{}, fmt.Errorf("save settings: %w", err)
		}
		return current, nil
	}
	s, err := domain.NewSettings(rawURL, clean)
	if err != nil {
		return domain.Settings{}, err
	}
	now := uc.clock.Now()
	cs, err := uc.source.Collections(ctx, s.Source, now.Add(-24*time.Hour), now.Add(horizon))
	if err != nil {
		return domain.Settings{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not load the waste calendar: {detail}", "detail", err.Error())
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return domain.Settings{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(cs, "", now)
	return s, nil
}

// saveRulesKeepingFile stores new rules for a file source without touching
// the stored file: Save would delete it.
func (uc *waste) saveRulesKeepingFile(ctx context.Context, s domain.Settings) error {
	return uc.repo.SaveWithFile(ctx, s, nil)
}

// UploadCalendar implements Waste. The file is parsed and validated before
// anything is stored; the file and the new source are saved together.
func (uc *waste) UploadCalendar(ctx context.Context, name string, content []byte) (domain.Settings, error) {
	if len(content) > domain.MaxCalendarBytes {
		return domain.Settings{}, kernel.NewValidationError("file", "The file is larger than 1 MB.")
	}
	if !utf8.Valid(content) {
		return domain.Settings{}, kernel.NewValidationError("file", "The file is not an ICS calendar (text/calendar).")
	}
	now := uc.clock.Now()
	all, err := uc.parser.Parse(content, now.Add(-fileWindow), now.Add(fileWindow))
	if err != nil {
		return domain.Settings{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "The file is not a valid ICS calendar: {detail}", "detail", err.Error())
	}
	src, err := domain.NewFileSource(name, now, all)
	if err != nil {
		return domain.Settings{}, err
	}
	current, err := uc.repo.Load(ctx)
	if err != nil && !errors.Is(err, domain.ErrNotConfigured) {
		return domain.Settings{}, fmt.Errorf("load settings: %w", err)
	}
	s := domain.Settings{Source: src, Rules: current.Rules}
	if s.Rules == nil {
		s.Rules = []domain.Rule{}
	}
	if err := uc.repo.SaveWithFile(ctx, s, content); err != nil {
		return domain.Settings{}, fmt.Errorf("save calendar: %w", err)
	}
	uc.store(inWindow(all, now.Add(-24*time.Hour), now.Add(horizon)), "", now)
	return s, nil
}

// RemoveCalendar implements Waste.
func (uc *waste) RemoveCalendar(ctx context.Context) (domain.Settings, error) {
	current, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return domain.Settings{Rules: []domain.Rule{}}, nil
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("load settings: %w", err)
	}
	if current.Source.Kind != domain.SourceFile {
		return current, nil
	}
	current.Source = domain.Source{}
	if err := uc.repo.Save(ctx, current); err != nil {
		return domain.Settings{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(nil, "", uc.clock.Now())
	return current, nil
}

// Upcoming implements Waste.
func (uc *waste) Upcoming(ctx context.Context) (Overview, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return Overview{}, err
	}
	if !s.Source.Configured() {
		return Overview{}, domain.ErrNotConfigured
	}
	uc.mu.RLock()
	loaded := uc.loaded
	uc.mu.RUnlock()
	if !loaded {
		_ = uc.Refresh(ctx) // failures are reported through Overview.Error
	}
	now := uc.clock.Now()
	o := Overview{}
	if s.Source.ExpiresWithin(now, domain.ExpiryWarning) {
		o.EndsOn = s.Source.File.To
	}
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	o.Types = domain.NextCollections(slices.Clone(uc.cs), s.Rules, now)
	o.UpdatedAt, o.Error = uc.updatedAt, uc.lastError
	return o, nil
}

// Refresh implements Waste; it is run by the background job. A file source
// re-reads the stored file and makes no outbound request.
func (uc *waste) Refresh(ctx context.Context) error {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	if !s.Source.Configured() {
		return nil
	}
	now := uc.clock.Now()
	cs, err := uc.source.Collections(ctx, s.Source, now.Add(-24*time.Hour), now.Add(horizon))
	if err != nil {
		uc.mu.Lock()
		uc.loaded, uc.lastError = true, "The waste calendar could not be loaded; showing the last known dates."
		uc.mu.Unlock()
		return fmt.Errorf("load waste calendar: %w", err)
	}
	uc.store(cs, "", now)
	return nil
}

func (uc *waste) store(cs []domain.Collection, errMsg string, at time.Time) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.loaded, uc.cs, uc.lastError, uc.updatedAt = true, cs, errMsg, at
}

func inWindow(cs []domain.Collection, from, to time.Time) []domain.Collection {
	from, to = domain.Date(from), domain.Date(to)
	out := make([]domain.Collection, 0, len(cs))
	for _, c := range cs {
		if !c.Date.Before(from) && !c.Date.After(to) {
			out = append(out, c)
		}
	}
	return out
}
