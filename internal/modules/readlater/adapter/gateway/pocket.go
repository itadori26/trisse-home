package gateway

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/readlater/domain"
)

// Codec implements usecase.ExportCodec.
type Codec struct{}

// Pocket implements usecase.ExportCodec: Pocket's CSV export has the header
// title,url,time_added,tags,status (status "unread" or "archive"); the
// columns are found by name, so their order does not matter.
func (Codec) Pocket(data []byte) ([]domain.ImportEntry, error) {
	r := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))))
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	urlCol, ok := col["url"]
	if !ok {
		return nil, errors.New("no url column")
	}
	field := func(rec []string, name string) string {
		if i, ok := col[name]; ok && i < len(rec) {
			return strings.TrimSpace(rec[i])
		}
		return ""
	}
	var out []domain.ImportEntry
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		if urlCol >= len(rec) || strings.TrimSpace(rec[urlCol]) == "" {
			continue
		}
		e := domain.ImportEntry{URL: strings.TrimSpace(rec[urlCol]), Title: field(rec, "title"), Archived: field(rec, "status") == "archive"}
		if sec, err := strconv.ParseInt(field(rec, "time_added"), 10, 64); err == nil && sec > 0 {
			e.Added = time.Unix(sec, 0).UTC()
		}
		out = append(out, e)
	}
}
