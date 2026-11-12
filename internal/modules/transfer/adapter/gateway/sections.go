// Package gateway adapts the modules' Portable implementations to the
// transfer use cases.
//
// Ring: Interface Adapters.
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/transfer/usecase"
	"trisse-home/internal/shared/kernel"
	"trisse-home/internal/shared/sqlutil"
)

// PortableSection implements usecase.Section for a modules.Portable.
type PortableSection struct{ P modules.Portable }

// Export implements usecase.Section.
func (s PortableSection) Export(ctx context.Context) ([]byte, error) {
	data, err := s.P.ExportData(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

// Import implements usecase.Section. Numbers are decoded exactly so IDs and
// integers survive the round trip.
func (s PortableSection) Import(ctx context.Context, raw []byte) error {
	var data map[string]sqlutil.Rows
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&data); err != nil {
		return fmt.Errorf("the section is not valid: %w", err)
	}
	if err := s.P.ImportData(ctx, data); err != nil {
		if errors.Is(err, sqlutil.ErrInvalidExport) {
			return kernel.NewPublicErrorf(kernel.ErrInvalid, "The data could not be imported: {detail}", "detail", err.Error())
		}
		return err
	}
	return nil
}

// Sections wraps the registry's portable modules.
func Sections(portables func() map[string]modules.Portable) func() map[string]usecase.Section {
	return func() map[string]usecase.Section {
		out := map[string]usecase.Section{}
		for name, p := range portables() {
			out[name] = PortableSection{P: p}
		}
		return out
	}
}
