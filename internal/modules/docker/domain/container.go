package domain

import (
	"slices"
	"strings"
	"time"
)

// Health of a container as reported by its health check.
type Health string

// Health values.
const (
	HealthNone      Health = "none"
	HealthHealthy   Health = "healthy"
	HealthUnhealthy Health = "unhealthy"
	HealthStarting  Health = "starting"
)

// Container is a Docker container as shown on the dashboard.
type Container struct {
	ID      string
	Name    string
	Image   string
	State   string // running, exited, restarting, paused, created, dead
	Status  string // human readable, e.g. "Up 3 hours (healthy)"
	Project string // compose project, if any
	Created time.Time
	// Labels are the container labels (used by service discovery, #78).
	Labels map[string]string
}

// Health derives the health check state from the status text.
func (c Container) Health() Health {
	switch {
	case strings.Contains(c.Status, "(healthy)"):
		return HealthHealthy
	case strings.Contains(c.Status, "(unhealthy)"):
		return HealthUnhealthy
	case strings.Contains(c.Status, "(health: starting)"):
		return HealthStarting
	default:
		return HealthNone
	}
}

// NeedsAttention reports whether the container is not running normally.
func (c Container) NeedsAttention() bool {
	return c.State != "running" || c.Health() == HealthUnhealthy
}

// Summary counts containers by condition.
type Summary struct {
	Total, Running, Stopped, Unhealthy int
}

// Summarise counts the containers.
func Summarise(cs []Container) Summary {
	s := Summary{Total: len(cs)}
	for _, c := range cs {
		switch {
		case c.State == "running" && c.Health() == HealthUnhealthy:
			s.Running++
			s.Unhealthy++
		case c.State == "running":
			s.Running++
		default:
			s.Stopped++
		}
	}
	return s
}

// SortForDisplay puts containers needing attention first, then sorts by name.
func SortForDisplay(cs []Container) {
	slices.SortStableFunc(cs, func(a, b Container) int {
		if a.NeedsAttention() != b.NeedsAttention() {
			if a.NeedsAttention() {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})
}
