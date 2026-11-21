package grafana

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/grafana/domain"
	"waylight/internal/modules/grafana/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Grafana connection (#610); the check is the
// search SaveSettings makes.
func connectionSpecs(client usecase.Client) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "grafana", Title: "Grafana",
		Fields: []modules.CredentialField{{Name: "token", Secret: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["token"])
			if err != nil {
				return err
			}
			if _, err := client.Search(ctx, s, ""); err != nil {
				return kernel.SetupFailure("Grafana", err)
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
	return []modules.SavedConnection{{Source: "grafana", Service: "grafana", URL: s.URL, Credential: map[string]string{"token": s.Token}}}, nil
}

// connections adapts the integrations resolver to usecase.Connections.
type connections struct{ r modules.ConnectionResolver }

// Settings implements usecase.Connections.
func (c connections) Settings(ctx context.Context, id string) (domain.Settings, error) {
	rc, err := c.r.ResolveConnection(ctx, "grafana", id)
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.NewSettings(rc.URL, rc.Credential["token"])
}
