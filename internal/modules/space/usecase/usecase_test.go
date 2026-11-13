package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/space/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ s domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error)   { return m.s, nil }
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = s; return nil }

type fakeLaunches struct {
	calls int
	err   error
	now   time.Time
}

func (f *fakeLaunches) Upcoming(context.Context, int) ([]domain.Launch, error) {
	f.calls++
	return []domain.Launch{{Mission: "past", Net: f.now.Add(-2 * time.Hour)}, {Mission: "A", Net: f.now.Add(time.Hour)}, {Mission: "B", Net: f.now.Add(48 * time.Hour)}}, f.err
}

type fakePictures struct {
	calls, imageCalls int
	key               string
	pic               domain.Picture
	err               error
}

func (f *fakePictures) Today(_ context.Context, key string) (domain.Picture, error) {
	f.calls++
	f.key = key
	return f.pic, f.err
}

func (f *fakePictures) Image(context.Context, string) (Image, error) {
	f.imageCalls++
	return Image{Bytes: []byte("jpg"), ContentType: "image/jpeg"}, nil
}

type clk struct{ now time.Time }

func (c *clk) Now() time.Time { return c.now }

func TestLaunches(t *testing.T) {
	ctx := context.Background()
	c := &clk{now: time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)}
	repo, l := &memRepo{}, &fakeLaunches{now: c.now}
	uc := New(repo, l, &fakePictures{}, c)
	if _, err := uc.Launches(ctx, 5); !errors.Is(err, kernel.ErrNotFound) || l.calls != 0 {
		t.Fatalf("disabled: %v %d", err, l.calls)
	}
	_, _ = uc.SaveSettings(ctx, true, false, "", false)
	got, err := uc.Launches(ctx, 1)
	if err != nil || len(got) != 1 || got[0].Mission != "A" {
		t.Fatalf("%+v %v", got, err)
	}
	// Cached for an hour; past launches drop out of the cached list.
	c.now = c.now.Add(59 * time.Minute)
	l.err = errors.New("rate limited")
	got, _ = uc.Launches(ctx, 5)
	if l.calls != 1 || len(got) != 2 {
		t.Fatalf("cache: %d %+v", l.calls, got)
	}
	c.now = c.now.Add(2 * time.Hour)
	if got, err := uc.Launches(ctx, 5); err != nil || l.calls != 2 || len(got) != 1 || got[0].Mission != "B" {
		t.Fatalf("stale on error: %+v %v", got, err)
	}
	fresh := New(&memRepo{s: domain.Settings{Launches: true}}, &fakeLaunches{err: errors.New("down")}, &fakePictures{}, c)
	if _, err := fresh.Launches(ctx, 5); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("%v", err)
	}
}

func TestPicture(t *testing.T) {
	ctx := context.Background()
	c := &clk{now: time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)}
	repo := &memRepo{}
	p := &fakePictures{pic: domain.Picture{Date: "2026-11-13", MediaType: "image", URL: "https://apod.nasa.gov/a.jpg"}}
	uc := New(repo, &fakeLaunches{}, p, c)
	if _, err := uc.Picture(ctx); !errors.Is(err, kernel.ErrNotFound) || p.calls != 0 {
		t.Fatal("disabled")
	}
	_, _ = uc.SaveSettings(ctx, false, true, "my-key", false)
	if pic, err := uc.Picture(ctx); err != nil || pic.Date != "2026-11-13" || p.key != "my-key" {
		t.Fatalf("%+v %v %q", pic, err, p.key)
	}
	for range 2 {
		if img, err := uc.PictureImage(ctx); err != nil || string(img.Bytes) != "jpg" {
			t.Fatalf("%+v %v", img, err)
		}
	}
	if p.calls != 1 || p.imageCalls != 1 {
		t.Fatalf("cached: %d %d", p.calls, p.imageCalls)
	}
	// Keeping the key does not reload; a video day has no image.
	_, _ = uc.SaveSettings(ctx, false, true, "", true)
	p.pic = domain.Picture{Date: "2026-11-14", MediaType: "video", URL: "https://youtube.example/x"}
	c.now = c.now.Add(4 * time.Hour)
	if _, err := uc.PictureImage(ctx); !errors.Is(err, ErrNoImage) {
		t.Fatalf("video: %v", err)
	}
}
