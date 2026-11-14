// Package presenter maps study modules and progress to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import "trisse-home/internal/modules/study/domain"

// ModuleDTO is one module.
type ModuleDTO struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	ECTS     float64  `json:"ects"`
	Semester int      `json:"semester"`
	Status   string   `json:"status" tstype:"'open' | 'in_progress' | 'passed'"`
	Grade    *float64 `json:"grade"`
}

// ProgressDTO is returned by GET /api/study/progress.
type ProgressDTO struct {
	TotalECTS      float64     `json:"totalEcts"`
	PassedECTS     float64     `json:"passedEcts"`
	InProgressECTS float64     `json:"inProgressEcts"`
	OpenECTS       float64     `json:"openEcts"`
	PassedModules  int         `json:"passedModules"`
	Modules        int         `json:"modules"`
	Percent        float64     `json:"percent"`
	Average        *float64    `json:"average"`
	GradedECTS     float64     `json:"gradedEcts"`
	InProgress     []ModuleDTO `json:"inProgress"`
}

// SettingsDTO is returned by the settings endpoints.
type SettingsDTO struct {
	TotalECTS float64 `json:"totalEcts"`
}

// Module maps a module.
func Module(m domain.Module) ModuleDTO {
	return ModuleDTO{ID: m.ID, Name: m.Name, ECTS: m.ECTS, Semester: m.Semester, Status: string(m.Status), Grade: m.Grade}
}

// Modules maps a list.
func Modules(ms []domain.Module) []ModuleDTO {
	out := make([]ModuleDTO, 0, len(ms))
	for _, m := range ms {
		out = append(out, Module(m))
	}
	return out
}

// Progress maps the progress and lists the modules in progress.
func Progress(p domain.Progress, ms []domain.Module) ProgressDTO {
	dto := ProgressDTO{TotalECTS: p.TotalECTS, PassedECTS: p.PassedECTS, InProgressECTS: p.InProgressECTS, OpenECTS: p.OpenECTS,
		PassedModules: p.PassedModules, Modules: p.Modules, Percent: p.Percent, Average: p.Average, GradedECTS: p.GradedECTS,
		InProgress: []ModuleDTO{}}
	for _, m := range ms {
		if m.Status == domain.StatusInProgress {
			dto.InProgress = append(dto.InProgress, Module(m))
		}
	}
	return dto
}
