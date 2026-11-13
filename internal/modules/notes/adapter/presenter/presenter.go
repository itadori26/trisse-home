package presenter

import (
	"time"

	"trisse-home/internal/modules/notes/domain"
	"trisse-home/internal/modules/notes/usecase"
)

// NoteDTO is a note in API responses.
type NoteDTO struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Pinned    bool      `json:"pinned"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Note maps a NoteView.
func Note(v usecase.NoteView) NoteDTO { return NoteDTO(v) }

// Notes maps a list of NoteViews.
func Notes(vs []usecase.NoteView) []NoteDTO {
	out := make([]NoteDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, Note(v))
	}
	return out
}

// SegmentDTO is a piece of highlighted search output; hit marks a match.
// Text is plain text, never HTML.
type SegmentDTO struct {
	Text string `json:"text"`
	Hit  bool   `json:"hit"`
}

// SearchHitDTO is a note found by GET /api/notes/search (#73).
type SearchHitDTO struct {
	ID      string       `json:"id"`
	Title   []SegmentDTO `json:"title"`
	Snippet []SegmentDTO `json:"snippet"`
}

// NoteRefDTO names a note (backlinks, #73).
type NoteRefDTO struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// RevisionDTO describes an earlier version of a note (#73).
type RevisionDTO struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	Size      int       `json:"size"`
}

// DiffLineDTO is one line of a diff: op " ", "+" or "-".
type DiffLineDTO struct {
	Op   string `json:"op"`
	Text string `json:"text"`
}

// RevisionDetailDTO is a revision with its body and the diff from the
// revision to the current note.
type RevisionDetailDTO struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	CreatedAt time.Time     `json:"createdAt"`
	Size      int           `json:"size"`
	Body      string        `json:"body"`
	Diff      []DiffLineDTO `json:"diff"`
}

func segmentsDTO(in []domain.Segment) []SegmentDTO {
	out := make([]SegmentDTO, 0, len(in))
	for _, s := range in {
		out = append(out, SegmentDTO(s))
	}
	return out
}

// SearchHits maps search results.
func SearchHits(hits []domain.SearchHit) []SearchHitDTO {
	out := make([]SearchHitDTO, 0, len(hits))
	for _, h := range hits {
		out = append(out, SearchHitDTO{ID: h.ID, Title: segmentsDTO(h.Title), Snippet: segmentsDTO(h.Snippet)})
	}
	return out
}

// NoteRefs maps backlinks.
func NoteRefs(refs []usecase.NoteRef) []NoteRefDTO {
	out := make([]NoteRefDTO, 0, len(refs))
	for _, r := range refs {
		out = append(out, NoteRefDTO(r))
	}
	return out
}

// Revisions maps revision summaries.
func Revisions(vs []usecase.RevisionView) []RevisionDTO {
	out := make([]RevisionDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, RevisionDTO(v))
	}
	return out
}

// RevisionDetail maps a revision with its diff.
func RevisionDetail(v usecase.RevisionDetail) RevisionDetailDTO {
	out := RevisionDetailDTO{ID: v.ID, Title: v.Title, CreatedAt: v.CreatedAt, Size: v.Size, Body: v.Body, Diff: make([]DiffLineDTO, 0, len(v.Diff))}
	for _, d := range v.Diff {
		out.Diff = append(out.Diff, DiffLineDTO{Op: string(d.Op), Text: d.Text})
	}
	return out
}
