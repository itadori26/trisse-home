package nextcloud

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/nextcloud/domain"
	"waylight/internal/modules/nextcloud/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Nextcloud connection (#610); the check is
// the status call SaveSettings makes.
func connectionSpecs(source usecase.StatusSource) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "nextcloud", Title: "Nextcloud",
		Fields: []modules.CredentialField{{Name: "username"}, {Name: "appPassword", Secret: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["username"], cred["appPassword"])
			if err != nil {
				return err
			}
			if _, err := source.Status(ctx, s); err != nil {
				return kernel.SetupFailure("Nextcloud", err)
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
	return []modules.SavedConnection{{Source: "nextcloud", Service: "nextcloud", URL: s.URL,
		Credential: map[string]string{"username": s.Username, "appPassword": s.AppPassword}}}, nil
}

// connections adapts the integrations resolver to usecase.Connections.
type connections struct{ r modules.ConnectionResolver }

// Settings implements usecase.Connections.
func (c connections) Settings(ctx context.Context, id string) (domain.Settings, error) {
	rc, err := c.r.ResolveConnection(ctx, "nextcloud", id)
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.NewSettings(rc.URL, rc.Credential["username"], rc.Credential["appPassword"])
}
