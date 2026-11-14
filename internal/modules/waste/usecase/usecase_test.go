package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	s    *domain.Settings
	file []byte
}

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s, m.file = &s, nil; return nil }
func (m *memRepo) SaveWithFile(_ context.Context, s domain.Settings, content []byte) error {
	m.s = &s
	if content != nil {
		m.file = content
	}
	return nil
}

type fakeSource struct {
	err   error
	calls []domain.SourceKind
}

// fakeParser returns dates for content "ok", none for "empty" and an error otherwise.
type fakeParser struct{}

func (fakeParser) Parse(content []byte, _, _ time.Time) ([]domain.Collection, error) {
	switch string(content) {
	case "ok":
		d := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
		return []domain.Collection{{Date: d, Summary: "Restmüll"}, {Date: d.AddDate(0, 0, 14), Summary: "Restmüll"}}, nil
	case "empty":
		return nil, nil
	default:
		return nil, errors.New("not an iCalendar object")
	}
}

func (f *fakeSource) Collections(_ context.Context, src domain.Source, _, _ time.Time) ([]domain.Collection, error) {
	f.calls = append(f.calls, src.Kind)
	d := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	return []domain.Collection{{Date: d, Summary: "Biotonne"}, {Date: d.AddDate(0, 0, 7), Summary: "Papier"}}, f.err
}

type clk struct{}

func (clk) Now() time.Time { return time.Date(2026, 11, 11, 19, 0, 0, 0, time.UTC) }

func TestWaste(t *testing.T) {
	ctx := context.Background()
	repo, src := &memRepo{}, &fakeSource{err: errors.New("404")}
	uc := NewWaste(repo, src, fakeParser{}, clk{})
	if _, err := uc.Upcoming(ctx); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("unconfigured: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, "https://example.org/a.ics", nil); err == nil || repo.s != nil {
		t.Fatal("broken subscription must not be saved")
	}
	src.err = nil
	if _, err := uc.SaveSettings(ctx, "https://example.org/a.ics", []domain.Rule{{Match: "bio", Label: "Organic"}}); err != nil {
		t.Fatal(err)
	}
	o, err := uc.Upcoming(ctx)
	if err != nil || len(o.Types) != 2 || o.Types[0].Style.Label != "Organic" || o.Types[0].DaysUntil != 1 {
		t.Fatalf("%+v %v", o, err)
	}
	src.err = errors.New("down")
	_ = uc.Refresh(ctx)
	if o, _ := uc.Upcoming(ctx); o.Error == "" || len(o.Types) != 2 {
		t.Fatalf("last known dates must stay: %+v", o)
	}
}

func TestUploadCalendar(t *testing.T) {
	ctx := context.Background()
	repo, src := &memRepo{}, &fakeSource{}
	uc := NewWaste(repo, src, fakeParser{}, clk{})
	if _, err := uc.SaveSettings(ctx, "https://example.org/a.ics", []domain.Rule{{Match: "rest", Label: "Black bin"}}); err != nil {
		t.Fatal(err)
	}
	for content, want := range map[string]error{"garbage": kernel.ErrInvalid, "empty": kernel.ErrInvalid, "\xff\xfe": kernel.ErrInvalid} {
		if _, err := uc.UploadCalendar(ctx, "a.ics", []byte(content)); !errors.Is(err, want) {
			t.Fatalf("%q: %v", content, err)
		}
		if repo.s.Source.Kind != domain.SourceURL || repo.file != nil {
			t.Fatalf("%q: invalid file changed the settings: %+v", content, repo.s)
		}
	}
	if _, err := uc.UploadCalendar(ctx, "a.ics", make([]byte, domain.MaxCalendarBytes+1)); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("oversize: %v", err)
	}

	s, err := uc.UploadCalendar(ctx, `C:\\Downloads\\abfall.ics`, []byte("ok"))
	if err != nil {
		t.Fatal(err)
	}
	f := s.Source.File
	if s.Source.Kind != domain.SourceFile || f.Name != "abfall.ics" || f.Events != 2 || f.From.Day() != 20 || f.To.Day() != 4 ||
		string(repo.file) != "ok" || s.Rules[0].Label != "Black bin" {
		t.Fatalf("upload: %+v file=%q", s, repo.file)
	}
	// The dates are there at once; no fetch was needed.
	src.calls = nil
	o, err := uc.Upcoming(ctx)
	if err != nil || len(o.Types) != 1 || o.Types[0].Style.Label != "Black bin" || len(src.calls) != 0 {
		t.Fatalf("%+v %v calls=%v", o, err, src.calls)
	}
	// Ends on 4 December, within 30 days of 11 November: the widget asks for next year's file.
	if o.EndsOn.IsZero() || o.EndsOn.Day() != 4 {
		t.Fatalf("expiry hint: %+v", o.EndsOn)
	}
	// Refresh re-reads the stored file through the source, without a URL.
	if err := uc.Refresh(ctx); err != nil || len(src.calls) != 1 || src.calls[0] != domain.SourceFile {
		t.Fatalf("refresh: %v %v", err, src.calls)
	}

	// Saving rules without a URL keeps the file.
	s, err = uc.SaveSettings(ctx, "", []domain.Rule{{Match: "rest", Label: "Grey bin"}})
	if err != nil || s.Source.Kind != domain.SourceFile || string(repo.file) != "ok" || repo.s.Rules[0].Label != "Grey bin" {
		t.Fatalf("rules only: %+v %v", s, err)
	}
	// Removing the file resets the source and keeps the rules.
	s, err = uc.RemoveCalendar(ctx)
	if err != nil || s.Source.Configured() || repo.file != nil || len(s.Rules) != 1 {
		t.Fatalf("remove: %+v %v", s, err)
	}
	if _, configured, _ := uc.Settings(ctx); configured {
		t.Fatal("removed source must read as not configured")
	}
	if _, err := uc.Upcoming(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("upcoming after remove: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, "", nil); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("rules without any source: %v", err)
	}
}
