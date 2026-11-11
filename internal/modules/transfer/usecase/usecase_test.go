package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/transfer/domain"
	"trisse-home/internal/shared/kernel"
)

type memSection struct {
	data []byte
	fail bool
}

func (m *memSection) Export(context.Context) ([]byte, error) { return m.data, nil }
func (m *memSection) Import(_ context.Context, d []byte) error {
	if m.fail {
		return errors.New("broken references")
	}
	m.data = d
	return nil
}

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

func TestTransfer(t *testing.T) {
	ctx := context.Background()
	notes, todos := &memSection{data: []byte(`{"notes":[]}`)}, &memSection{fail: true}
	uc := NewTransfer(func() map[string]Section { return map[string]Section{"notes": notes, "todos": todos} }, clk{})
	doc, err := uc.Export(ctx)
	if err != nil || doc.Format != domain.Format || len(doc.Sections) != 2 {
		t.Fatalf("%+v %v", doc, err)
	}
	doc.Sections["notes"] = []byte(`{"notes":[{"id":"n"}]}`)
	doc.Sections["calendar"] = []byte(`{}`)
	res, err := uc.Import(ctx, doc)
	if err != nil || len(res.Imported) != 1 || res.Skipped[0] != "calendar" || res.Failed["todos"] == "" || string(notes.data) != `{"notes":[{"id":"n"}]}` {
		t.Fatalf("%+v %v", res, err)
	}
	doc.Version = domain.Version + 1
	if _, err := uc.Import(ctx, doc); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("newer version accepted: %v", err)
	}
}
