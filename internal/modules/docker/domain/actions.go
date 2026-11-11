package domain

import (
	"regexp"
	"slices"

	"trisse-home/internal/shared/kernel"
)

// MaxLogLines is how many log lines are shown for a container.
const MaxLogLines = 200

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`)

// ErrActionNotAllowed is returned when actions are disabled or the container
// is not on the allow-list.
var ErrActionNotAllowed = kernel.NewPublicError(kernel.ErrForbidden,
	"Actions are not enabled for this container (DOCKER_ACTIONS / DOCKER_ACTIONS_CONTAINERS).")

// ActionPolicy decides which containers may be restarted and have their logs
// read. Actions are off unless enabled and every container must be listed
// explicitly by name.
type ActionPolicy struct {
	Enabled bool
	Allowed []string
}

// Allows reports whether actions are permitted for the container name.
func (p ActionPolicy) Allows(name string) bool {
	return p.Enabled && namePattern.MatchString(name) && slices.Contains(p.Allowed, name)
}

// Check returns ErrActionNotAllowed unless Allows(name).
func (p ActionPolicy) Check(name string) error {
	if !p.Allows(name) {
		return ErrActionNotAllowed
	}
	return nil
}

// AllowedNames lists the containers with actions (empty when disabled).
func (p ActionPolicy) AllowedNames() []string {
	if !p.Enabled {
		return []string{}
	}
	return slices.Clone(p.Allowed)
}
