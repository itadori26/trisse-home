package fritzbox

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/fritzbox/domain"
	"waylight/internal/modules/fritzbox/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the FRITZ!Box connection (#610). The login is
// optional: the internet connection is readable without it, the call list
// needs it. The check makes the same calls as SaveSettings.
func connectionSpecs(box usecase.Box) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "fritzbox", Title: "FRITZ!Box",
		Fields: []modules.CredentialField{{Name: "username", Optional: true}, {Name: "password", Secret: true, Optional: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["username"], cred["password"])
			if err != nil {
				return err
			}
			if _, _, err := box.WAN(ctx, s); err != nil {
				return kernel.SetupFailure("FRITZ!Box", err)
			}
			if s.HasLogin() {
				if _, err := box.Calls(ctx, s, 1); err != nil {
					return kernel.SetupFailure("FRITZ!Box", err)
				}
			}
			return nil
		},
	}}
}

// savedConnections returns the module settings as a connection to copy.
func savedConnections(ctx context.Context, repo usecase.SettingsRepository) ([]modules.SavedConnection, error) {
	s, err := repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []modules.SavedConnection{{Source: "fritzbox", Service: "fritzbox", URL: s.URL,
		Credential: map[string]string{"username": s.Username, "password": s.Password}}}, nil
}

// connections adapts the integrations resolver to usecase.Connections.
type connections struct{ r modules.ConnectionResolver }

// Settings implements usecase.Connections.
func (c connections) Settings(ctx context.Context, id string) (domain.Settings, error) {
	rc, err := c.r.ResolveConnection(ctx, "fritzbox", id)
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.NewSettings(rc.URL, rc.Credential["username"], rc.Credential["password"])
}
