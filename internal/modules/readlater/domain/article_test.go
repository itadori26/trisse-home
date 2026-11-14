package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestArticle(t *testing.T) {
	a, err := NewArticle("a", " https://Example.org/post#comments ", time.Unix(0, 0))
	if err != nil || a.URL != "https://Example.org/post" || a.Status != StatusUnread || a.Site != "Example.org" {
		t.Fatalf("%+v %v", a, err)
	}
	for _, bad := range []string{"", "ftp://x/y", "javascript:alert(1)", "https://u:p@x/y", "example.org", "http://169.254.169.254/latest"} {
		if _, err := NewArticle("a", bad, time.Now()); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%q: %v", bad, err)
		}
	}
	a.ApplyPage("  A   title  ", "Example", strings.Repeat("x ", 400), "<p>x</p>", 450)
	if a.Title != "A title" || a.ReadingMinutes() != 3 || len([]rune(a.Excerpt)) > 300 {
		t.Fatalf("%+v %d", a, a.ReadingMinutes())
	}
	if (Article{}).ReadingMinutes() != 1 || CountWords(" a  b\nc ") != 3 {
		t.Fatal("reading time")
	}
	if s, ok := ParseStatus("archived"); !ok || s != StatusArchived {
		t.Fatal("status")
	}
	if _, ok := ParseStatus("deleted"); ok {
		t.Fatal("unknown status")
	}
}

func TestResolveImages(t *testing.T) {
	in := `<p>a</p><img src="cid:0" alt="first"><p>b</p><img src="cid:1" alt=""><img src="cid:12">`
	got := ResolveImages(in, map[int]string{0: "/api/readlater/articles/x/images/0", 12: "/i/12"})
	want := `<p>a</p><img src="/api/readlater/articles/x/images/0" alt="first"><p>b</p><img src="/i/12">`
	if got != want {
		t.Fatalf("\n got %s\nwant %s", got, want)
	}
}
