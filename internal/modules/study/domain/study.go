package domain

import (
	"cmp"
	"math"
	"slices"
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Status of a study module.
type Status string

// Statuses.
const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusPassed     Status = "passed"
)

// Limits.
const (
	MaxModules       = 200
	maxNameRunes     = 80
	MaxECTS          = 60
	MaxSemester      = 20
	BestGrade        = 1.0
	WorstGrade       = 5.0
	DefaultTotalECTS = 180
	MaxTotalECTS     = 600
)

// ErrNotFound is returned for an unknown module.
var ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The module does not exist.")

// Module is one course or exam with its credits.
type Module struct {
	ID   string
	Name string
	// ECTS in half credits steps (e.g. 7.5).
	ECTS float64
	// Semester is 1–20, 0 = not planned yet.
	Semester int
	Status   Status
	// Grade on the German scale 1.0 (best) – 4.0 (pass); only for passed
	// modules, nil for ungraded ones (pass/fail).
	Grade *float64
}

// ModuleInput is the editable form of a module.
type ModuleInput struct {
	Name     string
	ECTS     float64
	Semester int
	Status   Status
	Grade    *float64
}

// NewModule validates input.
func NewModule(id string, in ModuleInput) (Module, error) {
	m := Module{ID: id, Name: strings.TrimSpace(in.Name), ECTS: in.ECTS, Semester: in.Semester, Status: in.Status}
	switch {
	case m.Name == "" || utf8.RuneCountInString(m.Name) > maxNameRunes:
		return Module{}, kernel.NewValidationError("name", "Enter a module name of at most 80 characters.")
	case math.IsNaN(m.ECTS) || m.ECTS <= 0 || m.ECTS > MaxECTS || m.ECTS*2 != math.Trunc(m.ECTS*2):
		return Module{}, kernel.NewValidationError("ects", "ECTS must be between 0.5 and 60 in steps of 0.5.")
	case m.Semester < 0 || m.Semester > MaxSemester:
		return Module{}, kernel.NewValidationError("semester", "The semester must be between 1 and 20 (or empty).")
	case m.Status != StatusOpen && m.Status != StatusInProgress && m.Status != StatusPassed:
		return Module{}, kernel.NewValidationError("status", "Choose open, in progress or passed.")
	}
	if in.Grade != nil {
		g := math.Round(*in.Grade*10) / 10
		if m.Status != StatusPassed {
			return Module{}, kernel.NewValidationError("grade", "Only passed modules have a grade.")
		}
		if math.IsNaN(*in.Grade) || g < BestGrade || g > 4.0 {
			return Module{}, kernel.NewValidationError("grade", "A passing grade lies between 1.0 and 4.0.")
		}
		m.Grade = &g
	}
	return m, nil
}

// Progress of the studies.
type Progress struct {
	TotalECTS      float64
	PassedECTS     float64
	InProgressECTS float64
	OpenECTS       float64
	PassedModules  int
	Modules        int
	// Percent of TotalECTS passed, 0–100.
	Percent float64
	// Average is the ECTS-weighted grade of the graded passed modules, nil
	// while there is none; GradedECTS are the credits it is based on.
	Average    *float64
	GradedECTS float64
}

// ComputeProgress sums up the modules against total ECTS.
func ComputeProgress(ms []Module, total float64) Progress {
	p := Progress{TotalECTS: total, Modules: len(ms)}
	var weighted float64
	for _, m := range ms {
		switch m.Status {
		case StatusPassed:
			p.PassedECTS += m.ECTS
			p.PassedModules++
			if m.Grade != nil {
				weighted += *m.Grade * m.ECTS
				p.GradedECTS += m.ECTS
			}
		case StatusInProgress:
			p.InProgressECTS += m.ECTS
		default:
			p.OpenECTS += m.ECTS
		}
	}
	if total > 0 {
		p.Percent = math.Min(100, p.PassedECTS/total*100)
	}
	if p.GradedECTS > 0 {
		avg := weighted / p.GradedECTS
		p.Average = &avg
	}
	return p
}

// Sort orders modules by semester (unplanned last), then name.
func Sort(ms []Module) {
	slices.SortStableFunc(ms, func(a, b Module) int {
		sa, sb := a.Semester, b.Semester
		if sa == 0 {
			sa = MaxSemester + 1
		}
		if sb == 0 {
			sb = MaxSemester + 1
		}
		return cmp.Or(cmp.Compare(sa, sb), strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)))
	})
}

// NewTotalECTS validates the credits of the whole degree.
func NewTotalECTS(v float64) (float64, error) {
	if math.IsNaN(v) || v < 1 || v > MaxTotalECTS {
		return 0, kernel.NewValidationError("totalEcts", "The total must be between 1 and 600 ECTS.")
	}
	return v, nil
}
