package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/system/domain"
	"trisse-home/internal/shared/kernel"
)

// updateCheckInterval is the minimum time between two release lookups.
const updateCheckInterval = 24 * time.Hour

// Release is the latest published release.
type Release struct {
	Version     string
	URL         string
	PublishedAt time.Time
}

// VersionInfo is returned by Versions.Info.
type VersionInfo struct {
	Current string
	// InstallMethod tells how updates reach this instance.
	InstallMethod   domain.InstallMethod
	CheckEnabled    bool
	Latest          *Release
	UpdateAvailable bool
	CheckedAt       time.Time
}

// Versions is the input port for the running version and the update check.
type Versions interface {
	Info(ctx context.Context) (VersionInfo, error)
	SetCheckEnabled(ctx context.Context, enabled bool) (VersionInfo, error)
	// Check refreshes the latest release when the check is enabled (job).
	Check(ctx context.Context) error
}

// ReleaseSource looks up the latest release (output port).
type ReleaseSource interface {
	Latest(ctx context.Context) (Release, bool, error)
}

// UpdateSettings persists whether the update check is enabled (output port).
type UpdateSettings interface {
	CheckEnabled(ctx context.Context) (bool, error)
	SetCheckEnabled(ctx context.Context, enabled bool) error
}

type versions struct {
	current  domain.Version
	method   domain.InstallMethod
	source   ReleaseSource
	settings UpdateSettings
	clock    kernel.Clock

	mu        sync.Mutex
	latest    *Release
	checkedAt time.Time
}

// NewVersions returns the version use cases for the running version current,
// installed with method.
func NewVersions(current string, method domain.InstallMethod, source ReleaseSource, settings UpdateSettings, clock kernel.Clock) Versions {
	return &versions{current: domain.ParseVersion(current), method: method, source: source, settings: settings, clock: clock}
}

// Info implements Versions (cache-aside: looks up at most once a day).
func (uc *versions) Info(ctx context.Context) (VersionInfo, error) {
	enabled, err := uc.settings.CheckEnabled(ctx)
	if err != nil {
		return VersionInfo{}, fmt.Errorf("load update settings: %w", err)
	}
	if enabled {
		uc.mu.Lock()
		stale := uc.clock.Now().Sub(uc.checkedAt) >= updateCheckInterval
		uc.mu.Unlock()
		if stale {
			_ = uc.Check(ctx) // a failed lookup only means no hint
		}
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	info := VersionInfo{Current: uc.current.String(), InstallMethod: uc.method, CheckEnabled: enabled, CheckedAt: uc.checkedAt}
	if enabled && uc.latest != nil {
		l := *uc.latest
		info.Latest = &l
		info.UpdateAvailable = domain.ParseVersion(l.Version).IsNewerThan(uc.current)
	}
	return info, nil
}

// SetCheckEnabled implements Versions.
func (uc *versions) SetCheckEnabled(ctx context.Context, enabled bool) (VersionInfo, error) {
	if err := uc.settings.SetCheckEnabled(ctx, enabled); err != nil {
		return VersionInfo{}, fmt.Errorf("save update settings: %w", err)
	}
	if !enabled {
		uc.mu.Lock()
		uc.latest, uc.checkedAt = nil, time.Time{}
		uc.mu.Unlock()
	}
	return uc.Info(ctx)
}

// Check implements Versions.
func (uc *versions) Check(ctx context.Context) error {
	enabled, err := uc.settings.CheckEnabled(ctx)
	if err != nil || !enabled {
		return err
	}
	rel, ok, err := uc.source.Latest(ctx)
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.checkedAt = uc.clock.Now()
	if err != nil {
		return fmt.Errorf("look up latest release: %w", err)
	}
	if ok {
		uc.latest = &rel
	}
	return nil
}
