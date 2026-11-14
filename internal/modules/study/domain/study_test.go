package domain

import (
	"errors"
	"math"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func g(v float64) *float64 { return &v }

func TestNewModule(t *testing.T) {
	m, err := NewModule("a", ModuleInput{Name: " Analysis I ", ECTS: 7.5, Semester: 1, Status: StatusPassed, Grade: g(1.34)})
	if err != nil || m.Name != "Analysis I" || *m.Grade != 1.3 {
		t.Fatalf("%+v %v", m, err)
	}
	for name, in := range map[string]ModuleInput{
		"no name":        {ECTS: 5, Status: StatusOpen},
		"zero ects":      {Name: "x", ECTS: 0, Status: StatusOpen},
		"quarter ects":   {Name: "x", ECTS: 2.25, Status: StatusOpen},
		"too many ects":  {Name: "x", ECTS: 61, Status: StatusOpen},
		"semester":       {Name: "x", ECTS: 5, Semester: 21, Status: StatusOpen},
		"status":         {Name: "x", ECTS: 5, Status: "failed"},
		"grade open":     {Name: "x", ECTS: 5, Status: StatusOpen, Grade: g(2)},
		"failing grade":  {Name: "x", ECTS: 5, Status: StatusPassed, Grade: g(5)},
		"too good grade": {Name: "x", ECTS: 5, Status: StatusPassed, Grade: g(0.7)},
		"nan grade":      {Name: "x", ECTS: 5, Status: StatusPassed, Grade: g(math.NaN())},
	} {
		if _, err := NewModule("a", in); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	// Pass/fail modules have no grade.
	if m, err := NewModule("a", ModuleInput{Name: "Praktikum", ECTS: 15, Status: StatusPassed}); err != nil || m.Grade != nil {
		t.Fatalf("%+v %v", m, err)
	}
}

func TestComputeProgress(t *testing.T) {
	ms := []Module{
		{Name: "A", ECTS: 10, Status: StatusPassed, Grade: g(1.0)},
		{Name: "B", ECTS: 5, Status: StatusPassed, Grade: g(2.5)},
		{Name: "Praktikum", ECTS: 15, Status: StatusPassed},
		{Name: "C", ECTS: 6, Status: StatusInProgress},
		{Name: "D", ECTS: 4, Status: StatusOpen},
	}
	p := ComputeProgress(ms, 180)
	if p.PassedECTS != 30 || p.InProgressECTS != 6 || p.OpenECTS != 4 || p.PassedModules != 3 || p.Modules != 5 || p.GradedECTS != 15 {
		t.Fatalf("%+v", p)
	}
	// (1.0·10 + 2.5·5) / 15 = 1.5
	if p.Average == nil || math.Abs(*p.Average-1.5) > 1e-9 || math.Abs(p.Percent-100.0/6) > 1e-9 {
		t.Fatalf("average %v percent %v", p.Average, p.Percent)
	}
	if p := ComputeProgress(nil, 180); p.Average != nil || p.Percent != 0 {
		t.Fatalf("empty: %+v", p)
	}
	if p := ComputeProgress([]Module{{ECTS: 200, Status: StatusPassed}}, 180); p.Percent != 100 {
		t.Fatalf("capped: %v", p.Percent)
	}
}

func TestSortAndTotal(t *testing.T) {
	ms := []Module{{Name: "b", Semester: 0}, {Name: "Z", Semester: 1}, {Name: "a", Semester: 2}, {Name: "A", Semester: 1}}
	Sort(ms)
	if ms[0].Name != "A" || ms[1].Name != "Z" || ms[2].Name != "a" || ms[3].Name != "b" {
		t.Fatalf("%+v", ms)
	}
	if _, err := NewTotalECTS(0); err == nil {
		t.Fatal("0")
	}
	if v, err := NewTotalECTS(210); err != nil || v != 210 {
		t.Fatal("210")
	}
}
