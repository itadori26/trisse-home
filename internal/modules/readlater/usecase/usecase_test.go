package usecase

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/readlater/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	m      map[string]domain.Article
	images map[string]map[int]Image
	extra  int64
}

func newRepo() *memRepo {
	return &memRepo{m: map[string]domain.Article{}, images: map[string]map[int]Image{}}
}
func (r *memRepo) List(context.Context) ([]domain.Article, error) {
	var out []domain.Article
	for _, a := range r.m {
		out = append(out, a)
	}
	return out, nil
}
func (r *memRepo) Get(_ context.Context, id string) (domain.Article, error) {
	a, ok := r.m[id]
	if !ok {
		return a, domain.ErrNotFound
	}
	return a, nil
}
func (r *memRepo) Save(_ context.Context, a domain.Article, images map[int]Image) error {
	r.m[a.ID], r.images[a.ID] = a, images
	return nil
}
func (r *memRepo) SetStatus(_ context.Context, id string, s domain.Status) error {
	a := r.m[id]
	a.Status = s
	r.m[id] = a
	return nil
}
func (r *memRepo) Delete(_ context.Context, id string) error {
	delete(r.m, id)
	delete(r.images, id)
	return nil
}
func (r *memRepo) Image(_ context.Context, id string, n int) (Image, error) {
	img, ok := r.images[id][n]
	if !ok {
		return Image{}, domain.ErrNotFound
	}
	return img, nil
}
func (r *memRepo) TotalBytes(context.Context) (int64, error) {
	t := r.extra
	for _, a := range r.m {
		t += a.Bytes
	}
	return t, nil
}

type fakePages struct {
	page Page
	err  error
}

func (f *fakePages) Read(context.Context, string) (Page, error) { return f.page, f.err }

type fakeImages struct {
	size  int
	fails map[string]bool
	calls int
}

func (f *fakeImages) Fetch(_ context.Context, url string) (Image, error) {
	f.calls++
	if f.fails[url] {
		return Image{}, errors.New("404")
	}
	return Image{Bytes: make([]byte, f.size), ContentType: "image/png"}, nil
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { c.t = c.t.Add(time.Minute); return c.t }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "a" + strconv.Itoa(i.n) }

func TestSave(t *testing.T) {
	ctx := context.Background()
	repo := newRepo()
	pages := &fakePages{page: Page{Title: "Post", Site: "Blog", Content: `<p>x</p><img src="cid:0" alt="a"><img src="cid:1" alt="b">`,
		Images: []string{"https://b.example/0.png", "https://b.example/1.png"}, Words: 500}}
	imgs := &fakeImages{size: 1000, fails: map[string]bool{"https://b.example/1.png": true}}
	uc := New(repo, pages, imgs, &clk{t: time.Unix(0, 0)}, &ids{})

	a, err := uc.Save(ctx, "https://b.example/post#top")
	if err != nil || a.Title != "Post" || a.ReadingMinutes() != 3 || a.Error != "" {
		t.Fatalf("%+v %v", a, err)
	}
	if a.Content != `<p>x</p><img src="/api/readlater/a1/images/0" alt="a">` || a.Bytes != int64(len(a.Content))+1000 {
		t.Fatalf("images not resolved: %q %d", a.Content, a.Bytes)
	}
	if img, err := uc.Image(ctx, "a1", 0); err != nil || len(img.Bytes) != 1000 {
		t.Fatalf("image: %v", err)
	}
	// Saving the same page again returns the existing article.
	if again, _ := uc.Save(ctx, "https://b.example/post"); again.ID != "a1" || len(repo.m) != 1 {
		t.Fatalf("duplicate: %+v", again)
	}

	// An unreachable page keeps the link with the reason.
	pages.err = kernel.NewPublicError(kernel.ErrInvalid, "The server answered 404 Not Found.")
	b, err := uc.Save(ctx, "https://b.example/gone")
	if err != nil || b.Content != "" || !strings.Contains(b.Error, "404") {
		t.Fatalf("%+v %v", b, err)
	}
}

func TestLimits(t *testing.T) {
	ctx := context.Background()
	repo := newRepo()
	var content strings.Builder
	var srcs []string
	for i := range 30 {
		content.WriteString(`<img src="cid:` + strconv.Itoa(i) + `" alt="">`)
		srcs = append(srcs, "https://b.example/"+strconv.Itoa(i))
	}
	imgs := &fakeImages{size: domain.MaxImageBytes}
	uc := New(repo, &fakePages{page: Page{Content: content.String(), Images: srcs}}, imgs, &clk{}, &ids{})
	a, err := uc.Save(ctx, "https://b.example/many")
	if err != nil {
		t.Fatal(err)
	}
	// 8 MB per article: four 2 MB images.
	if got := strings.Count(a.Content, "<img"); got != 4 || len(repo.images[a.ID]) != 4 {
		t.Fatalf("images kept: %d", got)
	}
	repo.extra = domain.MaxStorageBytes
	if _, err := uc.Save(ctx, "https://b.example/full"); !errors.Is(err, kernel.ErrConflict) {
		t.Fatalf("storage full: %v", err)
	}
}

func TestStatesAndList(t *testing.T) {
	ctx := context.Background()
	repo := newRepo()
	uc := New(repo, &fakePages{page: Page{Content: "<p>x</p>"}}, &fakeImages{}, &clk{t: time.Unix(0, 0)}, &ids{})
	a, _ := uc.Save(ctx, "https://x.example/1")
	b, _ := uc.Save(ctx, "https://x.example/2")
	if err := uc.SetStatus(ctx, a.ID, domain.StatusArchived); err != nil {
		t.Fatal(err)
	}
	if err := uc.SetStatus(ctx, a.ID, "gone"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
	if list, _ := uc.List(ctx, ""); len(list) != 1 || list[0].ID != b.ID || list[0].Content != "" {
		t.Fatalf("not archived: %+v", list)
	}
	if list, _ := uc.List(ctx, domain.StatusArchived); len(list) != 1 || list[0].ID != a.ID {
		t.Fatalf("archived: %+v", list)
	}
	if err := uc.Delete(ctx, a.ID); err != nil || len(repo.m) != 1 {
		t.Fatal(err)
	}
	if err := uc.Delete(ctx, a.ID); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal(err)
	}
}
