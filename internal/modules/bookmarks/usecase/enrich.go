package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/shared/kernel"
)

// maxChecksPerRun spreads the weekly link check over the hourly job runs.
const maxChecksPerRun = 25

var errEnrichDisabled = kernel.NewValidationError("settings", "Enable fetching titles and icons in the bookmark settings first.")
var errCheckDisabled = kernel.NewValidationError("settings", "Enable the link check in the bookmark settings first.")

func (d Deps) settings(ctx context.Context) (domain.Settings, error) {
	if d.Settings == nil {
		return domain.Settings{}, nil
	}
	return d.Settings.LoadSettings(ctx)
}

func (d Deps) canEnrich(s domain.Settings) bool {
	return s.Enrich && d.Metadata != nil && d.Icons != nil
}

func (d Deps) canCheck(s domain.Settings) bool {
	return s.CheckLinks && d.Checker != nil && d.Links != nil
}

// hostTitle reports whether the title is only the fallback host name, i.e.
// the user left it empty.
func hostTitle(b domain.Bookmark) bool {
	u, err := url.Parse(b.URL())
	return err == nil && b.Title() == u.Host
}

// enrich fetches title and favicon; errors are returned but never undo
// the saved bookmark.
func (d Deps) enrich(ctx context.Context, b domain.Bookmark) (domain.Bookmark, bool, error) {
	meta, err := d.Metadata.Fetch(ctx, b.URL())
	if err != nil {
		return b, false, err
	}
	if t := domain.CleanTitle(meta.Title); t != "" && hostTitle(b) {
		in := domain.BookmarkInput{FolderID: b.FolderID(), Title: t, URL: b.URL(), Description: b.Description(), Tags: b.Tags()}
		if err := b.Update(in); err == nil {
			if err := d.Repo.SaveBookmark(ctx, b); err != nil {
				return b, false, fmt.Errorf("save bookmark: %w", err)
			}
		}
	}
	icon, err := d.Metadata.FetchIcon(ctx, domain.ResolveIconURL(b.URL(), meta.IconURL))
	if err != nil {
		return b, false, nil // pages without a usable icon are fine
	}
	if err := d.Icons.SaveIcon(ctx, b.ID(), icon); err != nil {
		return b, false, fmt.Errorf("save icon: %w", err)
	}
	return b, true, nil
}

type getSettings struct{ d Deps }

// NewGetSettings returns the GetSettings use case.
func NewGetSettings(d Deps) GetSettings { return &getSettings{d: d} }

// Execute implements GetSettings.
func (uc *getSettings) Execute(ctx context.Context) (SettingsView, error) {
	s, err := uc.d.settings(ctx)
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView(s), nil
}

type saveSettings struct{ d Deps }

// NewSaveSettings returns the SaveSettings use case.
func NewSaveSettings(d Deps) SaveSettings { return &saveSettings{d: d} }

// Execute implements SaveSettings.
func (uc *saveSettings) Execute(ctx context.Context, v SettingsView) (SettingsView, error) {
	if uc.d.Settings == nil {
		return SettingsView{}, errors.New("settings are not available")
	}
	if err := uc.d.Settings.SaveSettings(ctx, domain.Settings(v)); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	return v, nil
}

type enrichBookmark struct{ d Deps }

// NewEnrichBookmark returns the EnrichBookmark use case.
func NewEnrichBookmark(d Deps) EnrichBookmark { return &enrichBookmark{d: d} }

// Execute implements EnrichBookmark.
func (uc *enrichBookmark) Execute(ctx context.Context, id string) (BookmarkView, error) {
	b, err := uc.d.Repo.GetBookmark(ctx, id)
	if err != nil {
		return BookmarkView{}, err
	}
	s, err := uc.d.settings(ctx)
	if err != nil {
		return BookmarkView{}, fmt.Errorf("load settings: %w", err)
	}
	if !uc.d.canEnrich(s) {
		return BookmarkView{}, errEnrichDisabled
	}
	b, _, err = uc.d.enrich(ctx, b)
	if err != nil {
		return BookmarkView{}, kernel.NewPublicError(kernel.ErrUnavailable, "The page could not be read.")
	}
	return uc.d.viewOne(ctx, b)
}

type checkLink struct{ d Deps }

// NewCheckLink returns the CheckLink use case.
func NewCheckLink(d Deps) CheckLink { return &checkLink{d: d} }

// Execute implements CheckLink.
func (uc *checkLink) Execute(ctx context.Context, id string) (BookmarkView, error) {
	b, err := uc.d.Repo.GetBookmark(ctx, id)
	if err != nil {
		return BookmarkView{}, err
	}
	s, err := uc.d.settings(ctx)
	if err != nil {
		return BookmarkView{}, fmt.Errorf("load settings: %w", err)
	}
	if !uc.d.canCheck(s) {
		return BookmarkView{}, errCheckDisabled
	}
	if err := uc.d.check(ctx, b); err != nil {
		return BookmarkView{}, err
	}
	return uc.d.viewOne(ctx, b)
}

func (d Deps) check(ctx context.Context, b domain.Bookmark) error {
	code, final, failed := d.Checker.Check(ctx, b.URL())
	status := domain.ClassifyLink(code, b.URL(), final, failed, d.Clock.Now())
	if err := d.Links.SaveLinkStatus(ctx, b.ID(), status); err != nil {
		return fmt.Errorf("save link status: %w", err)
	}
	return nil
}

type checkLinks struct{ d Deps }

// NewCheckLinks returns the CheckLinks use case.
func NewCheckLinks(d Deps) CheckLinks { return &checkLinks{d: d} }

// Execute implements CheckLinks: the links checked longest ago first, at
// most maxChecksPerRun per run.
func (uc *checkLinks) Execute(ctx context.Context) (int, error) {
	s, err := uc.d.settings(ctx)
	if err != nil {
		return 0, fmt.Errorf("load settings: %w", err)
	}
	if !uc.d.canCheck(s) {
		return 0, nil
	}
	all, err := uc.d.Repo.ListBookmarks(ctx)
	if err != nil {
		return 0, fmt.Errorf("list bookmarks: %w", err)
	}
	statuses, err := uc.d.Links.ListLinkStatuses(ctx)
	if err != nil {
		return 0, fmt.Errorf("list link statuses: %w", err)
	}
	now := uc.d.Clock.Now()
	var due []domain.Bookmark
	for _, b := range all {
		if statuses[b.ID()].Due(now) {
			due = append(due, b)
		}
	}
	slices.SortStableFunc(due, func(a, b domain.Bookmark) int {
		return statuses[a.ID()].CheckedAt.Compare(statuses[b.ID()].CheckedAt)
	})
	checked := 0
	for _, b := range due[:min(len(due), maxChecksPerRun)] {
		if ctx.Err() != nil {
			break
		}
		if err := uc.d.check(ctx, b); err != nil {
			return checked, err
		}
		checked++
	}
	return checked, nil
}

type getIcon struct{ d Deps }

// NewGetIcon returns the GetIcon use case.
func NewGetIcon(d Deps) GetIcon { return &getIcon{d: d} }

// Execute implements GetIcon.
func (uc *getIcon) Execute(ctx context.Context, id string) (domain.Icon, error) {
	if uc.d.Icons == nil {
		return domain.Icon{}, domain.ErrIconNotFound
	}
	return uc.d.Icons.Icon(ctx, id)
}

// viewOne returns a bookmark with its link status and icon flag.
func (d Deps) viewOne(ctx context.Context, b domain.Bookmark) (BookmarkView, error) {
	v := toView(b)
	if d.Links != nil {
		statuses, err := d.Links.ListLinkStatuses(ctx)
		if err != nil {
			return BookmarkView{}, fmt.Errorf("list link statuses: %w", err)
		}
		v.Link = statuses[b.ID()]
	}
	if d.Icons != nil {
		icons, err := d.Icons.IconIDs(ctx)
		if err != nil {
			return BookmarkView{}, fmt.Errorf("list icons: %w", err)
		}
		v.HasIcon = icons[b.ID()]
	}
	return v, nil
}
