package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewFeed(t *testing.T) {
	if _, err := NewFeed("1", "file:///etc/passwd", time.Now()); err == nil {
		t.Fatal("non-http URL accepted")
	}
	f, err := NewFeed("1", " https://go.dev/blog/feed.atom ", time.Now())
	if err != nil || f.Title != f.URL {
		t.Fatalf("unexpected %+v %v", f, err)
	}
	f.RecordSuccess("The Go Blog", "javascript:alert(1)", time.Now())
	if f.Title != "The Go Blog" || f.SiteURL != "" {
		t.Fatalf("metadata not sanitised: %+v", f)
	}
	f.RecordFailure(errors.New("timeout"), time.Now())
	if f.LastError != "timeout" {
		t.Fatal("error not recorded")
	}
}

func TestNewItem(t *testing.T) {
	pub := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	it := NewItem("1", "f", "", "  Hello \n  world ", "javascript:alert(1)", pub)
	if it.Title != "Hello world" || it.Link != "" || it.GUID == "" {
		t.Fatalf("not sanitised: %+v", it)
	}
	it = NewItem("1", "f", "", "", "https://x.y/post", pub)
	if it.GUID != "https://x.y/post" || it.Title != "(untitled)" {
		t.Fatalf("fallbacks wrong: %+v", it)
	}
}
