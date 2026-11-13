package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"trisse-home/internal/modules/releases/domain"
	"trisse-home/internal/shared/kernel"
)

const (
	radiohead = "a74b1b7f-71a5-4011-9441-d0b5e4122711"
	bjork     = "87c5dedd-371d-4a53-9f7f-80522fb7f3cb"
	album     = "11111111-1111-1111-1111-111111111111"
)

var now = time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)

type memRepo struct{ list []domain.Artist }

func (m *memRepo) List(context.Context) ([]domain.Artist, error) { return slices.Clone(m.list), nil }
func (m *memRepo) Add(_ context.Context, a domain.Artist) error {
	m.list = append(m.list, a)
	return nil
}
func (m *memRepo) Remove(_ context.Context, id string) error {
	m.list = slices.DeleteFunc(m.list, func(a domain.Artist) bool { return a.ID == id })
	return nil
}

type fakeMB struct {
	down    map[string]bool
	calls   int
	queries []string
}

func (f *fakeMB) SearchArtists(_ context.Context, q string) ([]domain.Artist, error) {
	f.queries = append(f.queries, q)
	return []domain.Artist{{ID: radiohead, Name: "Radiohead"}}, nil
}

func (f *fakeMB) ReleasesSince(_ context.Context, a domain.Artist, since time.Time) ([]domain.Release, error) {
	f.calls++
	if f.down[a.ID] {
		return nil, errors.New("503")
	}
	if !since.Equal(now.AddDate(0, 0, -7*domain.MaxWeeks)) {
		return nil, errors.New("unexpected period")
	}
	return []domain.Release{{ID: a.ID, Title: a.Name + " LP", Artist: a.Name, Type: "Album", Date: now.AddDate(0, 0, -2)}}, nil
}

func (f *fakeMB) Cover(context.Context, string) (domain.Cover, error) {
	return domain.Cover{ContentType: "image/jpeg", Data: []byte{1}}, nil
}

type clk struct{}

func (clk) Now() time.Time { return now }

func TestFollowAndList(t *testing.T) {
	ctx := context.Background()
	repo, mb := &memRepo{}, &fakeMB{down: map[string]bool{}}
	uc := NewReleases(repo, mb, clk{})

	if list, _ := uc.SearchArtists(ctx, " r "); len(list) != 0 || len(mb.queries) != 0 {
		t.Fatal("single letters are not searched")
	}
	if list, err := uc.SearchArtists(ctx, "radiohead"); err != nil || len(list) != 1 {
		t.Fatalf("search: %v", err)
	}
	if v, err := uc.List(ctx, 4); err != nil || len(v.Releases) != 0 {
		t.Fatalf("empty list without artists: %+v %v", v, err)
	}
	if err := uc.Follow(ctx, radiohead, "Radiohead"); err != nil {
		t.Fatal(err)
	}
	if err := uc.Follow(ctx, radiohead, "Radiohead"); err != nil || len(repo.list) != 1 {
		t.Fatalf("following twice is a no-op: %v", err)
	}
	if v, _ := uc.List(ctx, 4); len(v.Releases) != 1 || v.Releases[0].Title != "Radiohead LP" {
		t.Fatalf("a new artist's releases appear at once: %+v", v)
	}
	if err := uc.Follow(ctx, "nope", "x"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("invalid id: %v", err)
	}

	_ = uc.Follow(ctx, bjork, "Björk")
	mb.down[bjork] = true
	if err := uc.Refresh(ctx); err == nil {
		t.Fatal("a partial failure is reported to the log")
	}
	v, err := uc.List(ctx, 4)
	if err != nil || len(v.Releases) != 1 || v.Error == "" {
		t.Fatalf("other artists stay visible: %+v %v", v, err)
	}
	if err := uc.Unfollow(ctx, radiohead); err != nil {
		t.Fatal(err)
	}
	if v, _ := uc.List(ctx, 4); len(v.Releases) != 0 {
		t.Fatalf("unfollowing removes the releases: %+v", v.Releases)
	}
}

func TestLimitAndCover(t *testing.T) {
	ctx := context.Background()
	repo, mb := &memRepo{}, &fakeMB{down: map[string]bool{}}
	uc := NewReleases(repo, mb, clk{})
	for range domain.MaxFollowed {
		repo.list = append(repo.list, domain.Artist{ID: "x"})
	}
	if err := uc.Follow(ctx, bjork, "Björk"); !errors.Is(err, domain.ErrTooManyArtists) {
		t.Fatalf("limit: %v", err)
	}
	repo.list = []domain.Artist{{ID: radiohead, Name: "Radiohead"}}
	_ = uc.Refresh(ctx)
	if _, err := uc.Cover(ctx, radiohead); err != nil {
		t.Fatalf("cover of a listed release: %v", err)
	}
	for _, id := range []string{album, "../x"} {
		if _, err := uc.Cover(ctx, id); !errors.Is(err, domain.ErrCoverNotFound) {
			t.Errorf("%q: only listed releases are proxied, got %v", id, err)
		}
	}
}
