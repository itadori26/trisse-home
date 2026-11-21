package smarthome

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/modules/smarthome/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Home Assistant connection (#610); the check
// reads the states like SaveSettings.
func connectionSpecs(provider usecase.Provider) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "homeassistant", Title: "Home Assistant",
		Fields: []modules.CredentialField{{Name: "token", Secret: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["token"])
			if err != nil {
				return err
			}
			if _, err := provider.States(ctx, s); err != nil {
				return kernel.SetupFailure("Home Assistant", err)
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
	// A connection from signing in (#654) holds a short-lived access token
	// that only this module renews; copied, it would stop working within
	// half an hour.
	if s.Refresh != "" {
		return nil, nil
	}
	return []modules.SavedConnection{{Source: "smarthome", Service: "homeassistant", URL: s.URL, Credential: map[string]string{"token": s.Token}}}, nil
}

// connections adapts the integrations resolver to usecase.Connections.
type connections struct{ r modules.ConnectionResolver }

// connectionsOf returns nil without a resolver, so the use cases only read
// the module settings.
func connectionsOf(r modules.ConnectionResolver) usecase.Connections {
	if r == nil {
		return nil
	}
	return connections{r}
}

// Settings implements usecase.Connections.
func (c connections) Settings(ctx context.Context, id string) (domain.Settings, error) {
	rc, err := c.r.ResolveConnection(ctx, "homeassistant", id)
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.NewSettings(rc.URL, rc.Credential["token"])
}
