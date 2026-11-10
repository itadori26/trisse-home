package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"trisse-home/internal/modules/bookmarks/domain"
)

type memRepo struct {
	bookmarks map[string]domain.Bookmark
	folders   map[string]domain.Folder
}

func newMem() *memRepo {
	return &memRepo{bookmarks: map[string]domain.Bookmark{}, folders: map[string]domain.Folder{}}
}

func (m *memRepo) ListBookmarks(context.Context) ([]domain.Bookmark, error) {
	var out []domain.Bookmark
	for _, b := range m.bookmarks {
		out = append(out, b)
	}
	return out, nil
}
func (m *memRepo) GetBookmark(_ context.Context, id string) (domain.Bookmark, error) {
	b, ok := m.bookmarks[id]
	if !ok {
		return b, domain.ErrBookmarkNotFound
	}
	return b, nil
}
func (m *memRepo) SaveBookmark(_ context.Context, b domain.Bookmark) error {
	m.bookmarks[b.ID()] = b
	return nil
}
func (m *memRepo) DeleteBookmark(_ context.Context, id string) error {
	delete(m.bookmarks, id)
	return nil
}
func (m *memRepo) ListFolders(context.Context) ([]domain.Folder, error) {
	var out []domain.Folder
	for _, f := range m.folders {
		out = append(out, f)
	}
	return out, nil
}
func (m *memRepo) GetFolder(_ context.Context, id string) (domain.Folder, error) {
	f, ok := m.folders[id]
	if !ok {
		return f, domain.ErrFolderNotFound
	}
	return f, nil
}
func (m *memRepo) SaveFolder(_ context.Context, f domain.Folder) error {
	m.folders[f.ID()] = f
	return nil
}
func (m *memRepo) DeleteFolder(_ context.Context, id string) error {
	delete(m.folders, id)
	for k, b := range m.bookmarks {
		if b.FolderID() == id {
			b.MoveToFolder("")
			m.bookmarks[k] = b
		}
	}
	return nil
}

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return fmt.Sprint(i.n) }

func TestBookmarkUseCases(t *testing.T) {
	ctx := context.Background()
	repo := newMem()
	d := Deps{Repo: repo, Clock: clk{}, IDs: &ids{}}

	folder, err := NewCreateFolder(d).Execute(ctx, "Dev")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewCreateBookmark(d).Execute(ctx, Input{URL: "https://x.y", FolderID: "missing"}); !errors.Is(err, domain.ErrFolderNotFound) {
		t.Fatalf("expected folder not found, got %v", err)
	}
	gd, err := NewCreateBookmark(d).Execute(ctx, Input{URL: "https://go.dev", FolderID: folder.ID, Tags: []string{"Go"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewCreateBookmark(d).Execute(ctx, Input{URL: "https://svelte.dev", Title: "Svelte", Tags: []string{"web"}}); err != nil {
		t.Fatal(err)
	}

	list := NewListBookmarks(d)
	tests := []struct {
		name string
		f    Filter
		want int
	}{
		{"all", Filter{}, 2},
		{"folder", Filter{FolderID: folder.ID}, 1},
		{"unfiled", Filter{FolderID: UnfiledFolderID}, 1},
		{"tag", Filter{Tag: "GO"}, 1},
		{"query", Filter{Query: "svel"}, 1},
		{"nothing", Filter{Query: "rust"}, 0},
	}
	for _, tt := range tests {
		res, err := list.Execute(ctx, tt.f)
		if err != nil || len(res.Bookmarks) != tt.want {
			t.Errorf("%s: got %d bookmarks, want %d (%v)", tt.name, len(res.Bookmarks), tt.want, err)
		}
	}
	res, _ := list.Execute(ctx, Filter{})
	if len(res.Tags) != 2 || res.Folders[0].Count != 1 {
		t.Fatalf("unexpected aggregates %+v %+v", res.Tags, res.Folders)
	}

	if _, err := NewUpdateBookmark(d).Execute(ctx, gd.ID, Input{URL: "https://go.dev/doc", Title: "Go docs"}); err != nil {
		t.Fatal(err)
	}
	if err := NewDeleteFolder(d).Execute(ctx, folder.ID); err != nil {
		t.Fatal(err)
	}
	if b, _ := repo.GetBookmark(ctx, gd.ID); b.FolderID() != "" {
		t.Fatal("bookmark not unfiled after folder deletion")
	}
	if err := NewDeleteBookmark(d).Execute(ctx, "missing"); !errors.Is(err, domain.ErrBookmarkNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if _, err := NewRenameFolder(d).Execute(ctx, "missing", "x"); !errors.Is(err, domain.ErrFolderNotFound) {
		t.Fatalf("expected folder not found, got %v", err)
	}
}
