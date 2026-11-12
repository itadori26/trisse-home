package domain

import (
	"fmt"
	"time"

	"trisse-home/internal/shared/kernel"
)

// DefaultStaleAfter is the age after which an export is shown as unknown.
const DefaultStaleAfter = 2 * time.Hour

// ErrNoExport is returned when no export exists (yet).
var ErrNoExport = kernel.NewPublicError(kernel.ErrNotFound,
	"No S.M.A.R.T. export found. Run scripts/smart-export.sh on the host (see README).")

// Protocol of a disk.
type Protocol string

// Protocols; SCSI and others only get the overall verdict.
const (
	ProtocolATA   Protocol = "ATA"
	ProtocolNVMe  Protocol = "NVMe"
	ProtocolOther Protocol = "other"
)

// Status of a disk, ordered by severity.
type Status string

// Statuses.
const (
	StatusOK       Status = "ok"
	StatusUnknown  Status = "unknown"
	StatusWarning  Status = "warning"
	StatusCritical Status = "critical"
)

func rank(s Status) int {
	switch s {
	case StatusWarning:
		return 2
	case StatusCritical:
		return 3
	case StatusUnknown:
		return 1
	}
	return 0
}

// Worse returns the more severe status.
func Worse(a, b Status) Status {
	if rank(b) > rank(a) {
		return b
	}
	return a
}

// Attribute is one ATA S.M.A.R.T. attribute.
type Attribute struct {
	ID         int
	Name       string
	Value      int // normalised value
	Threshold  int
	Raw        int64
	FailingNow bool // when_failed == "now"
}

// NVMeHealth is the NVMe SMART / health information log.
type NVMeHealth struct {
	CriticalWarning int
	AvailableSpare  int
	SpareThreshold  int
	PercentageUsed  int
	MediaErrors     int64
}

// Report is the normalised smartctl output of one disk.
type Report struct {
	Device      string // e.g. /dev/sda
	Model       string
	Protocol    Protocol
	CapacityB   int64
	At          time.Time // when smartctl ran
	Passed      *bool     // overall self-assessment; nil when unavailable
	Temperature *int      // °C
	PowerOnH    *int64
	Attributes  []Attribute
	NVMe        *NVMeHealth
}

// Finding explains a status. Code and Value let clients translate it;
// Message is the English text (notifications, timeline).
type Finding struct {
	Status  Status
	Code    string
	Value   int64
	Name    string // attribute name for "attribute_failing"
	Message string
}

// Disk is an evaluated report.
type Disk struct {
	Report
	Status   Status
	Findings []Finding
	// Reallocated and Pending sectors (ATA ids 5 and 197), nil when unknown.
	Reallocated *int64
	Pending     *int64
}

// Well-known ATA attribute ids.
const (
	attrReallocated   = 5
	attrReportedUncor = 187
	attrPending       = 197
	attrOfflineUncor  = 198
)

var countAttributes = map[int]struct{ code, format string }{
	attrReallocated:   {"reallocated", "%d reallocated sectors"},
	attrReportedUncor: {"reported_uncorrectable", "%d uncorrectable errors reported"},
	attrPending:       {"pending", "%d sectors pending reallocation"},
	attrOfflineUncor:  {"offline_uncorrectable", "%d offline uncorrectable sectors"},
}

// Evaluate applies the health rules (ATA and NVMe) to a report.
func Evaluate(r Report, now time.Time, staleAfter time.Duration) Disk {
	d := Disk{Report: r, Status: StatusOK}
	add := func(f Finding) {
		d.Status = Worse(d.Status, f.Status)
		d.Findings = append(d.Findings, f)
	}
	if r.Passed != nil && !*r.Passed {
		add(Finding{Status: StatusCritical, Code: "self_assessment_failed", Message: "The disk's self-assessment failed"})
	}
	if r.Temperature != nil {
		warn, crit := 60, 70
		if r.Protocol == ProtocolNVMe {
			warn, crit = 70, 80
		}
		switch t := *r.Temperature; {
		case t >= crit:
			add(Finding{Status: StatusCritical, Code: "temperature", Value: int64(t), Message: fmt.Sprintf("Temperature %d °C", t)})
		case t >= warn:
			add(Finding{Status: StatusWarning, Code: "temperature", Value: int64(t), Message: fmt.Sprintf("Temperature %d °C", t)})
		}
	}
	for _, a := range r.Attributes {
		switch a.ID {
		case attrReallocated:
			d.Reallocated = &a.Raw
		case attrPending:
			d.Pending = &a.Raw
		}
		if a.FailingNow {
			add(Finding{Status: StatusCritical, Code: "attribute_failing", Name: a.Name, Message: fmt.Sprintf("Attribute %s is below its threshold", a.Name)})
			continue
		}
		if c, ok := countAttributes[a.ID]; ok && a.Raw > 0 {
			add(Finding{Status: StatusWarning, Code: c.code, Value: a.Raw, Message: fmt.Sprintf(c.format, a.Raw)})
		}
	}
	if n := r.NVMe; n != nil {
		if n.CriticalWarning != 0 {
			add(Finding{Status: StatusCritical, Code: "nvme_critical_warning", Value: int64(n.CriticalWarning), Message: fmt.Sprintf("NVMe critical warning 0x%02x", n.CriticalWarning)})
		}
		if n.SpareThreshold > 0 && n.AvailableSpare < n.SpareThreshold {
			add(Finding{Status: StatusCritical, Code: "spare_low", Value: int64(n.AvailableSpare), Message: fmt.Sprintf("Available spare %d %% is below the threshold of %d %%", n.AvailableSpare, n.SpareThreshold)})
		}
		switch {
		case n.PercentageUsed >= 100:
			add(Finding{Status: StatusCritical, Code: "wear", Value: int64(n.PercentageUsed), Message: fmt.Sprintf("%d %% of the rated endurance used", n.PercentageUsed)})
		case n.PercentageUsed >= 90:
			add(Finding{Status: StatusWarning, Code: "wear", Value: int64(n.PercentageUsed), Message: fmt.Sprintf("%d %% of the rated endurance used", n.PercentageUsed)})
		}
		if n.MediaErrors > 0 {
			add(Finding{Status: StatusWarning, Code: "media_errors", Value: n.MediaErrors, Message: fmt.Sprintf("%d media errors", n.MediaErrors)})
		}
	}
	if r.Passed == nil && r.NVMe == nil && len(r.Attributes) == 0 {
		add(Finding{Status: StatusUnknown, Code: "no_data", Message: "smartctl returned no health data"})
	}
	// A stale export says nothing about the current state.
	if staleAfter > 0 && now.Sub(r.At) > staleAfter {
		d.Status = StatusUnknown
		d.Findings = append([]Finding{{Status: StatusUnknown, Code: "stale", Message: "The export is outdated; check the cron job of scripts/smart-export.sh"}}, d.Findings...)
	}
	return d
}

// Observation turns an evaluated disk into a kernel observation; ok=false
// for unknown disks, which must neither alarm nor count as recovered.
func (d Disk) Observation() (kernel.Observation, bool) {
	if d.Status == StatusUnknown {
		return kernel.Observation{}, false
	}
	o := kernel.Observation{Kind: "diskhealth", Key: d.Device, Label: d.Label(), Healthy: d.Status == StatusOK}
	if len(d.Findings) > 0 {
		o.Detail = d.Findings[0].Message
	}
	return o, true
}

// Label names the disk for people.
func (d Disk) Label() string {
	if d.Model != "" {
		return d.Model + " (" + d.Device + ")"
	}
	return d.Device
}
