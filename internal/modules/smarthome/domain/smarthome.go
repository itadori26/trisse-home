package domain

import (
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// ErrNotConfigured is returned while no smart home hub is configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "Home Assistant is not configured yet.")

// ErrEntityNotFound is returned for an unknown entity ID.
var ErrEntityNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The entity does not exist.")

// ErrActionNotAllowed is returned when the entity is not on the allow-list.
var ErrActionNotAllowed = kernel.NewPublicError(kernel.ErrForbidden,
	"Actions are not enabled for this entity (SMARTHOME_ACTIONS_ENTITIES).")

// ErrNoAction is returned for entities that cannot be switched or pressed.
var ErrNoAction = kernel.NewPublicError(kernel.ErrInvalid, "This entity has no action; only switches and buttons can be used.")

// Settings are the connection details of Home Assistant.
type Settings struct {
	URL string
	// Token is a long-lived access token.
	Token string
}

// NewSettings validates and normalises the settings.
func NewSettings(rawURL, token string) (Settings, error) {
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Settings{}, kernel.NewValidationError("url", "The URL must look like http://homeassistant.local:8123.")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return Settings{}, kernel.NewValidationError("token", "A long-lived access token is required.")
	}
	return Settings{URL: rawURL, Token: token}, nil
}

// entityIDPattern matches Home Assistant entity IDs such as switch.coffee.
var entityIDPattern = regexp.MustCompile(`^[a-z0-9_]+\.[a-z0-9_]+$`)

// ValidEntityID reports whether id looks like an entity ID.
func ValidEntityID(id string) bool { return entityIDPattern.MatchString(id) }

// Entity is the state of one device or sensor.
type Entity struct {
	ID          string
	Name        string
	State       string
	Unit        string
	LastChanged time.Time
}

// Domain returns the part before the dot, e.g. "switch".
func (e Entity) Domain() string {
	d, _, _ := strings.Cut(e.ID, ".")
	return d
}

// ActionKind is what a tile can do with an entity.
type ActionKind string

// Action kinds.
const (
	ActionNone   ActionKind = ""
	ActionToggle ActionKind = "toggle"
	ActionPress  ActionKind = "press"
)

// Service is a Home Assistant service call, e.g. switch.toggle.
type Service struct {
	Domain, Name string
}

// services maps entity domains to their action; everything else is read-only.
var services = map[string]struct {
	kind    ActionKind
	service string
}{
	"switch":        {ActionToggle, "toggle"},
	"light":         {ActionToggle, "toggle"},
	"fan":           {ActionToggle, "toggle"},
	"input_boolean": {ActionToggle, "toggle"},
	"button":        {ActionPress, "press"},
	"input_button":  {ActionPress, "press"},
	"script":        {ActionPress, "turn_on"},
	"scene":         {ActionPress, "turn_on"},
}

// Action returns what can be done with the entity and the service to call.
func (e Entity) Action() (ActionKind, Service) {
	s, ok := services[e.Domain()]
	if !ok {
		return ActionNone, Service{}
	}
	return s.kind, Service{Domain: e.Domain(), Name: s.service}
}

// SortByName orders entities by name, then ID.
func SortByName(list []Entity) {
	slices.SortFunc(list, func(a, b Entity) int {
		if c := strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
}

// ActionPolicy lists the entities that may be switched or pressed. Actions
// are off while the list is empty; it comes from the environment, so a
// session alone can never widen it.
type ActionPolicy struct {
	Allowed []string
}

// Allows reports whether actions are permitted for the entity.
func (p ActionPolicy) Allows(id string) bool { return slices.Contains(p.Allowed, id) }

// Check returns the service to call for an allowed entity with an action.
func (p ActionPolicy) Check(e Entity) (Service, error) {
	if !p.Allows(e.ID) {
		return Service{}, ErrActionNotAllowed
	}
	kind, svc := e.Action()
	if kind == ActionNone {
		return Service{}, ErrNoAction
	}
	return svc, nil
}
