// Package presenter maps evaluated disks to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/diskhealth/domain"
)

// FindingDTO explains a status.
type FindingDTO struct {
	Status string `json:"status" tstype:"'ok' | 'unknown' | 'warning' | 'critical'"`
	// Code identifies the rule for translation, Value and Name fill it in.
	Code    string `json:"code"`
	Value   int64  `json:"value"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// DiskDTO is one disk. Optional values are null when smartctl did not report them.
type DiskDTO struct {
	Device        string       `json:"device"`
	Model         string       `json:"model"`
	Protocol      string       `json:"protocol" tstype:"'ATA' | 'NVMe' | 'other'"`
	CapacityBytes int64        `json:"capacityBytes"`
	CheckedAt     time.Time    `json:"checkedAt"`
	Status        string       `json:"status" tstype:"'ok' | 'unknown' | 'warning' | 'critical'"`
	Findings      []FindingDTO `json:"findings"`
	Temperature   *int         `json:"temperature"`
	PowerOnHours  *int64       `json:"powerOnHours"`
	Reallocated   *int64       `json:"reallocated"`
	Pending       *int64       `json:"pending"`
	// PercentageUsed is the NVMe wear indicator.
	PercentageUsed *int `json:"percentageUsed"`
}

// Disks maps evaluated disks.
func Disks(list []domain.Disk) []DiskDTO {
	out := make([]DiskDTO, 0, len(list))
	for _, d := range list {
		dto := DiskDTO{
			Device: d.Device, Model: d.Model, Protocol: string(d.Protocol), CapacityBytes: d.CapacityB, CheckedAt: d.At,
			Status: string(d.Status), Findings: make([]FindingDTO, 0, len(d.Findings)),
			Temperature: d.Temperature, PowerOnHours: d.PowerOnH, Reallocated: d.Reallocated, Pending: d.Pending,
		}
		for _, f := range d.Findings {
			dto.Findings = append(dto.Findings, FindingDTO{Status: string(f.Status), Code: f.Code, Value: f.Value, Name: f.Name, Message: f.Message})
		}
		if d.NVMe != nil {
			used := d.NVMe.PercentageUsed
			dto.PercentageUsed = &used
		}
		out = append(out, dto)
	}
	return out
}
