package domain

import (
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Label prefixes read from containers (#78). Homepage labels are accepted as
// a fallback so stacks prepared for gethomepage.dev work unchanged.
const (
	LabelPrefix         = "trisse-home."
	HomepageLabelPrefix = "homepage."
)

// Discovery interval limits in minutes.
const (
	DefaultDiscoveryInterval = 5
	MaxDiscoveryInterval     = 24 * 60
)

// ErrDiscoveryUnavailable is returned when the Docker integration is off.
var ErrDiscoveryUnavailable = kernel.NewPublicError(kernel.ErrInvalid, "Docker discovery needs the docker module to be enabled.")

// Container is what discovery needs to know about a container.
type Container struct {
	Name   string
	Labels map[string]string
}

// DiscoveredService is a service tile described by container labels.
type DiscoveredService struct {
	// Container is the container name; it identifies the tile across
	// container re-creations (the ID changes, the name does not).
	Container   string
	Name        string
	URL         string
	Icon        string
	Description string
}

// Validate checks the label values with the rules of a manual tile.
func (d DiscoveredService) Validate() error {
	_, err := NewService("", d.Name, d.URL, "", d.Icon, d.Description, false, 0)
	return err
}

// labelKeys maps a field to its trisse-home and Homepage label suffixes.
var labelKeys = struct{ name, url, icon, description [2]string }{
	name: [2]string{"name", "name"}, url: [2]string{"url", "href"},
	icon: [2]string{"icon", "icon"}, description: [2]string{"description", "description"},
}

// ParseLabels reads a discovered service from the labels of a container.
// It reports false when the container carries no usable URL label or opts
// out with trisse-home.enable=false. The name defaults to the container name.
func ParseLabels(c Container) (DiscoveredService, bool) {
	if strings.EqualFold(strings.TrimSpace(c.Labels[LabelPrefix+"enable"]), "false") {
		return DiscoveredService{}, false
	}
	get := func(keys [2]string) string {
		if v := strings.TrimSpace(c.Labels[LabelPrefix+keys[0]]); v != "" {
			return v
		}
		return strings.TrimSpace(c.Labels[HomepageLabelPrefix+keys[1]])
	}
	d := DiscoveredService{Container: c.Name, Name: get(labelKeys.name), URL: get(labelKeys.url),
		Icon: get(labelKeys.icon), Description: get(labelKeys.description)}
	if d.URL == "" || c.Name == "" {
		return DiscoveredService{}, false
	}
	if d.Name == "" {
		d.Name = c.Name
	}
	return d, true
}

// MergeRule decides who wins when a discovered tile was edited by hand.
type MergeRule string

// Merge rules.
const (
	// LabelsWin overwrites manual edits with the label values on every sync.
	LabelsWin MergeRule = "labels"
	// ManualWins keeps a tile as soon as the user edited it.
	ManualWins MergeRule = "manual"
)

// Valid reports whether r is a known rule.
func (r MergeRule) Valid() bool { return r == LabelsWin || r == ManualWins }

// DiscoverySettings configure the automatic sync.
type DiscoverySettings struct {
	Enabled bool
	// IntervalMinutes between automatic syncs.
	IntervalMinutes int
	Rule            MergeRule
}

// DefaultDiscoverySettings are used until the user saves settings; the
// sync is opt-in because it creates tiles on its own.
func DefaultDiscoverySettings() DiscoverySettings {
	return DiscoverySettings{IntervalMinutes: DefaultDiscoveryInterval, Rule: LabelsWin}
}

// Validate checks the settings.
func (s DiscoverySettings) Validate() error {
	if s.IntervalMinutes < 1 || s.IntervalMinutes > MaxDiscoveryInterval {
		return kernel.NewValidationErrorf("intervalMinutes", "The interval must be between 1 and {max} minutes.", "max", MaxDiscoveryInterval)
	}
	if !s.Rule.Valid() {
		return kernel.NewValidationError("rule", "Unknown merge rule.")
	}
	return nil
}

// Due reports whether an automatic sync should run now.
func (s DiscoverySettings) Due(last, now time.Time) bool {
	return s.Enabled && (last.IsZero() || now.Sub(last) >= time.Duration(s.IntervalMinutes)*time.Minute)
}

// SyncPlan lists the changes a sync applies.
type SyncPlan struct {
	Create []DiscoveredService
	// Update holds existing tiles with their new values.
	Update []Service
	// Hide and Show hold tiles whose container disappeared or came back.
	Hide, Show []Service
}

// Empty reports whether the plan changes nothing.
func (p SyncPlan) Empty() bool {
	return len(p.Create)+len(p.Update)+len(p.Hide)+len(p.Show) == 0
}

// PlanSync compares the discovered services with the existing tiles. Tiles
// are matched by container name; manual tiles (no container) are never
// touched. Tiles of vanished containers are hidden, not deleted, so their
// position, history and manual edits survive a container re-creation.
func PlanSync(existing []Service, found []DiscoveredService, rule MergeRule) SyncPlan {
	byContainer := make(map[string]Service, len(existing))
	for _, s := range existing {
		if s.Container != "" {
			byContainer[s.Container] = s
		}
	}
	var p SyncPlan
	seen := make(map[string]bool, len(found))
	for _, d := range found {
		if seen[d.Container] {
			continue
		}
		seen[d.Container] = true
		s, ok := byContainer[d.Container]
		if !ok {
			p.Create = append(p.Create, d)
			continue
		}
		if s.Hidden {
			s.Hidden = false
			p.Show = append(p.Show, s)
		}
		if rule == ManualWins && s.Edited {
			continue
		}
		if updated, changed := s.applyLabels(d); changed {
			p.Update = append(p.Update, updated)
		}
	}
	for _, s := range existing {
		if s.Container != "" && !s.Hidden && !seen[s.Container] {
			s.Hidden = true
			p.Hide = append(p.Hide, s)
		}
	}
	return p
}

// applyLabels copies the label values; the health check URL follows the URL
// unless it was set separately. Manual edits are forgotten (labels won).
func (s Service) applyLabels(d DiscoveredService) (Service, bool) {
	next := s
	if s.CheckURL == s.URL {
		next.CheckURL = d.URL
	}
	next.Name, next.URL, next.Icon, next.Description, next.Edited = d.Name, d.URL, d.Icon, d.Description, false
	return next, next != s
}
