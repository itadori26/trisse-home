package usecase

import (
	"context"
	"slices"

	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/shared/kernel"
)

// Compact implements SmartHome (#614): the name and state of one entity.
// A state is neither good nor bad, so the level stays none. The entity is
// returned too: the module turns its action into the request a client sends.
func (uc *smartHome) Compact(ctx context.Context, connectionID, id string) (kernel.CompactState, EntityView, error) {
	if id == "" {
		return kernel.CompactState{}, EntityView{}, kernel.NewValidationError("entity", "Choose one entity for the compact view.")
	}
	if !domain.ValidEntityID(id) {
		return kernel.CompactState{}, EntityView{}, domain.ErrEntityNotFound
	}
	states, err := uc.current(ctx, connectionID)
	if err != nil {
		return kernel.CompactState{}, EntityView{}, err
	}
	i := slices.IndexFunc(states, func(e domain.Entity) bool { return e.ID == id })
	if i < 0 {
		return kernel.CompactState{}, EntityView{}, domain.ErrEntityNotFound
	}
	e := uc.view(states[i])
	state := kernel.CompactState{
		Title:     kernel.NewTextf("{name}", "name", e.Name),
		Detail:    stateText(e.Entity),
		Level:     kernel.CompactNone,
		UpdatedAt: e.LastChanged,
	}
	if state.UpdatedAt.IsZero() {
		state.UpdatedAt = uc.d.Clock.Now()
	}
	return state, e, nil
}

// stateText names the common switch states so clients can translate them;
// sensor values keep their own text and unit.
func stateText(e domain.Entity) kernel.Text {
	switch e.State {
	case "on":
		return kernel.NewText("On")
	case "off":
		return kernel.NewText("Off")
	case "unavailable":
		return kernel.NewText("Unavailable")
	}
	if e.Unit != "" {
		return kernel.NewTextf("{state} {unit}", "state", e.State, "unit", e.Unit)
	}
	return kernel.NewTextf("{state}", "state", e.State)
}
