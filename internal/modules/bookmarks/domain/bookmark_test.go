package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNewBookmark(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		in      BookmarkInput
		wantErr bool
		check   func(Bookmark) bool
	}{
		{"title falls back to host", BookmarkInput{URL: "https://go.dev/doc"}, false, func(b Bookmark) bool { return b.Title() == "go.dev" }},
		{"tags normalised", BookmarkInput{URL: "https://x.y", Tags: []string{" Go ", "#go", "Server"}}, false,
			func(b Bookmark) bool { return strings.Join(b.Tags(), ",") == "go,server" }},
		{"umlaut tag", BookmarkInput{URL: "https://x.y", Tags: []string{"bücher"}}, false, nil},
		{"bad scheme", BookmarkInput{URL: "javascript:alert(1)"}, true, nil},
		{"bad tag", BookmarkInput{URL: "https://x.y", Tags: []string{"a b"}}, true, nil},
		{"too many tags", BookmarkInput{URL: "https://x.y", Tags: strings.Split("a,b,c,d,e,f,g,h,i,j,k", ",")}, true, nil},
		{"long title", BookmarkInput{URL: "https://x.y", Title: strings.Repeat("x", 201)}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBookmark("1", tt.in, now)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil && !tt.check(b) {
				t.Fatalf("check failed for %+v", b)
			}
		})
	}
}

func TestFolder(t *testing.T) {
	if _, err := NewFolder("1", "  ", 0); err == nil {
		t.Fatal("expected error for blank name")
	}
	f, err := NewFolder("1", " Dev ", 0)
	if err != nil || f.Name() != "Dev" {
		t.Fatalf("unexpected %v %v", f, err)
	}
}
