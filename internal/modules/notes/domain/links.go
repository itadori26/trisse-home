package domain

import (
	"regexp"
	"slices"
	"strings"
	"time"
)

var (
	wikiLink   = regexp.MustCompile(`\[\[([^\[\]|\n]+)(?:\|[^\[\]\n]*)?\]\]`)
	fencedCode = regexp.MustCompile("(?s)```.*?```|~~~.*?~~~")
	inlineCode = regexp.MustCompile("`[^`\n]*`")
)

// WikiLinks returns the note titles a Markdown body links to with
// [[Title]] or [[Title|label]], in order and without duplicates (#73).
// Links inside code are ignored.
func WikiLinks(body string) []string {
	body = fencedCode.ReplaceAllString(body, "")
	body = inlineCode.ReplaceAllString(body, "")
	var out []string
	for _, m := range wikiLink.FindAllStringSubmatch(body, -1) {
		title := strings.TrimSpace(m[1])
		if title != "" && !slices.ContainsFunc(out, func(t string) bool { return SameTitle(t, title) }) {
			out = append(out, title)
		}
	}
	return out
}

// SameTitle compares note titles the way links resolve them: trimmed and
// case-insensitive.
func SameTitle(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

// LinksTo reports whether the note links to a note with the given title.
func (n Note) LinksTo(title string) bool {
	return slices.ContainsFunc(WikiLinks(n.body), func(t string) bool { return SameTitle(t, title) })
}

// Revision is an earlier version of a note (#73).
type Revision struct {
	ID        string
	NoteID    string
	Title     string
	Body      string
	CreatedAt time.Time
}

// Revision limits: edits within RevisionGap form one revision, and at most
// MaxRevisions are kept per note.
const (
	RevisionGap  = 10 * time.Minute
	MaxRevisions = 50
)

// ContentChanged reports whether title or body differ from the note.
func (n Note) ContentChanged(title, body string) bool {
	return n.title != title || n.body != body
}

// ShouldRecordRevision decides whether saving now keeps the previous
// content as a revision: only when the content changed and the newest
// revision is older than RevisionGap, so a burst of autosaves while typing
// is one revision.
func ShouldRecordRevision(changed bool, newest, now time.Time) bool {
	return changed && (newest.IsZero() || now.Sub(newest) >= RevisionGap)
}

// RevisionsToPrune returns the IDs beyond MaxRevisions; revs are newest first.
func RevisionsToPrune(revs []Revision) []string {
	if len(revs) <= MaxRevisions {
		return nil
	}
	var out []string
	for _, r := range revs[MaxRevisions:] {
		out = append(out, r.ID)
	}
	return out
}

// DiffOp marks a line of a diff.
type DiffOp string

// Diff operations.
const (
	DiffSame   DiffOp = " "
	DiffAdd    DiffOp = "+"
	DiffRemove DiffOp = "-"
)

// DiffLine is one line of a line-based diff.
type DiffLine struct {
	Op   DiffOp
	Text string
}

// maxDiffLines bounds the quadratic diff; longer texts are shown as a full
// replacement.
const maxDiffLines = 3000

// DiffLines returns a line diff from a to b (longest common subsequence).
func DiffLines(a, b string) []DiffLine {
	x, y := strings.Split(a, "\n"), strings.Split(b, "\n")
	if len(x) > maxDiffLines || len(y) > maxDiffLines {
		out := make([]DiffLine, 0, len(x)+len(y))
		for _, l := range x {
			out = append(out, DiffLine{DiffRemove, l})
		}
		for _, l := range y {
			out = append(out, DiffLine{DiffAdd, l})
		}
		return out
	}
	// lcs[i][j] = length of the LCS of x[i:] and y[j:].
	lcs := make([][]int, len(x)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(y)+1)
	}
	for i := len(x) - 1; i >= 0; i-- {
		for j := len(y) - 1; j >= 0; j-- {
			if x[i] == y[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var out []DiffLine
	i, j := 0, 0
	for i < len(x) && j < len(y) {
		switch {
		case x[i] == y[j]:
			out = append(out, DiffLine{DiffSame, x[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, DiffLine{DiffRemove, x[i]})
			i++
		default:
			out = append(out, DiffLine{DiffAdd, y[j]})
			j++
		}
	}
	for ; i < len(x); i++ {
		out = append(out, DiffLine{DiffRemove, x[i]})
	}
	for ; j < len(y); j++ {
		out = append(out, DiffLine{DiffAdd, y[j]})
	}
	return out
}

// SearchQuery turns user input into a safe FTS5 query: every word becomes
// a quoted prefix term, all terms must match. Returns "" for no words.
func SearchQuery(input string) string {
	var terms []string
	for _, w := range strings.Fields(input) {
		w = strings.Map(func(r rune) rune {
			if r == '"' || r == '*' || r == '^' {
				return -1
			}
			return r
		}, w)
		if w != "" {
			terms = append(terms, `"`+w+`"*`)
		}
		if len(terms) == 10 {
			break
		}
	}
	return strings.Join(terms, " ")
}

// Segment is a piece of highlighted search output; Hit marks matches.
type Segment struct {
	Text string
	Hit  bool
}

// SearchHit is a note found by the full-text search with highlighted
// title and a highlighted excerpt of the body.
type SearchHit struct {
	ID      string
	Title   []Segment
	Snippet []Segment
}
