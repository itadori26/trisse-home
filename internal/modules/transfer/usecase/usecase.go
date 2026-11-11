package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	"trisse-home/internal/modules/transfer/domain"
	"trisse-home/internal/shared/kernel"
)

// Section is the exportable data of one module (output port).
type Section interface {
	Export(ctx context.Context) ([]byte, error)
	Import(ctx context.Context, data []byte) error
}

// Document is an export.
type Document struct {
	Format     string
	Version    int
	ExportedAt time.Time
	Sections   map[string][]byte
}

// ImportResult reports what happened per section.
type ImportResult struct {
	Imported []string
	// Skipped sections belong to unknown or disabled modules.
	Skipped []string
	// Failed maps section names to the reason (data stayed unchanged).
	Failed map[string]string
}

// Transfer is the input port of the module.
type Transfer interface {
	Export(ctx context.Context) (Document, error)
	Import(ctx context.Context, doc Document) (ImportResult, error)
}

type transfer struct {
	sections func() map[string]Section
	clock    kernel.Clock
}

// NewTransfer returns the use cases; sections is called on every request so
// it sees all registered modules.
func NewTransfer(sections func() map[string]Section, clock kernel.Clock) Transfer {
	return &transfer{sections: sections, clock: clock}
}

// Export implements Transfer.
func (uc *transfer) Export(ctx context.Context) (Document, error) {
	doc := Document{Format: domain.Format, Version: domain.Version, ExportedAt: uc.clock.Now(), Sections: map[string][]byte{}}
	for name, s := range uc.sections() {
		data, err := s.Export(ctx)
		if err != nil {
			return Document{}, fmt.Errorf("export %s: %w", name, err)
		}
		doc.Sections[name] = data
	}
	return doc, nil
}

// Import implements Transfer. Each section is imported atomically; a failing
// section keeps its data and is reported, the others are still imported.
func (uc *transfer) Import(ctx context.Context, doc Document) (ImportResult, error) {
	if err := domain.CheckHeader(doc.Format, doc.Version); err != nil {
		return ImportResult{}, err
	}
	available := uc.sections()
	res := ImportResult{Imported: []string{}, Skipped: []string{}, Failed: map[string]string{}}
	names := make([]string, 0, len(doc.Sections))
	for name := range doc.Sections {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		s, ok := available[name]
		if !ok {
			res.Skipped = append(res.Skipped, name)
			continue
		}
		if err := s.Import(ctx, doc.Sections[name]); err != nil {
			res.Failed[name] = err.Error()
			continue
		}
		res.Imported = append(res.Imported, name)
	}
	return res, nil
}
