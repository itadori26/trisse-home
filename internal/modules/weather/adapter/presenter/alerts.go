package presenter

import (
	"time"

	"trisse-home/internal/modules/weather/domain"
)

// WarningDTO is an official weather warning.
type WarningDTO struct {
	ID          string     `json:"id"`
	Event       string     `json:"event"`
	Headline    string     `json:"headline"`
	Description string     `json:"description"`
	Instruction string     `json:"instruction"`
	Severity    string     `json:"severity" tstype:"'minor' | 'moderate' | 'severe' | 'extreme'"`
	Onset       *time.Time `json:"onset"`
	Expires     *time.Time `json:"expires"`
}

func optionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// Warnings maps warnings.
func Warnings(ws []domain.Warning) []WarningDTO {
	out := make([]WarningDTO, 0, len(ws))
	for _, w := range ws {
		out = append(out, WarningDTO{ID: w.ID, Event: w.Event, Headline: w.Headline, Description: w.Description, Instruction: w.Instruction,
			Severity: string(w.Severity), Onset: optionalTime(w.Onset), Expires: optionalTime(w.Expires)})
	}
	return out
}
