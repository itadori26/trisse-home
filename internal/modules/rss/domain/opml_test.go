package domain

import (
	"strings"
	"testing"
)

func TestPlanImport(t *testing.T) {
	existing := []Feed{{URL: "https://Blog.Example.org/feed/"}}
	plan := PlanImport(existing, []ImportEntry{
		{Title: "Blog", URL: "https://blog.example.org/feed"}, // already subscribed (case, slash)
		{Title: "Go", URL: " https://go.dev/blog/feed.atom ", Category: "  Tech   /  Go "},
		{Title: "Go again", URL: "https://go.dev/blog/feed.atom#x"}, // repeated in the file
		{Title: "Bad", URL: "javascript:alert(1)"},
		{Title: "Site", URL: "http://news.example/rss", SiteURL: "ftp://x"},
	})
	if plan.Duplicates != 2 || len(plan.Add) != 2 || len(plan.Invalid) != 1 || plan.Invalid[0].Title != "Bad" {
		t.Fatalf("%+v", plan)
	}
	if plan.Add[0].URL != "https://go.dev/blog/feed.atom" || plan.Add[0].Category != "Tech / Go" || plan.Add[1].SiteURL != "" {
		t.Fatalf("%+v", plan.Add)
	}
	var many []ImportEntry
	for i := range MaxImportFeeds + 3 {
		many = append(many, ImportEntry{URL: "https://example.org/" + strings.Repeat("x", i%7) + "/" + string(rune('a'+i%26)) + "/" + itoa(i)})
	}
	if p := PlanImport(nil, many); len(p.Add) != MaxImportFeeds || len(p.Invalid) != 3 {
		t.Fatalf("limit: %d %d", len(p.Add), len(p.Invalid))
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for ; i > 0; i /= 10 {
		s = string(rune('0'+i%10)) + s
	}
	return s
}
