package media

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/media/domain"
	"waylight/internal/modules/media/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares one service per media server (#610), so a
// widget can pick a Jellyfin and a Plex connection side by side. The check
// reads the overview like SaveSettings.
func connectionSpecs(providers map[domain.Provider]usecase.MediaServerProvider) []modules.ConnectionSpec {
	out := make([]modules.ConnectionSpec, 0, len(providers))
	for _, p := range []domain.Provider{domain.ProviderJellyfin, domain.ProviderPlex} {
		provider := providers[p]
		out = append(out, modules.ConnectionSpec{
			Service: string(p), Title: p.Title(),
			Fields: []modules.CredentialField{{Name: "token", Secret: true}},
			Check: func(ctx context.Context, url string, cred map[string]string) error {
				s, err := domain.NewSettings(string(p), url, cred["token"])
				if err != nil {
					return err
				}
				if _, err := provider.Overview(ctx, s); err != nil {
					return kernel.SetupFailure(p.Title(), err)
				}
				return nil
			},
		})
	}
	return out
}

// savedConnections returns the media server of the module settings as a
// connection of its provider's service.
func savedConnections(ctx context.Context, repo usecase.SettingsRepository) ([]modules.SavedConnection, error) {
	s, err := repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []modules.SavedConnection{{Source: "media", Service: string(s.Provider), URL: s.URL, Credential: map[string]string{"token": s.Token}}}, nil
}

// connections adapts the integrations resolver to usecase.Connections; a
// connection may be of either media server.
type connections struct{ r modules.ConnectionResolver }

// Settings implements usecase.Connections.
func (c connections) Settings(ctx context.Context, id string) (domain.Settings, error) {
	for _, p := range []domain.Provider{domain.ProviderJellyfin, domain.ProviderPlex} {
		rc, err := c.r.ResolveConnection(ctx, string(p), id)
		if errors.Is(err, modules.ErrNoConnection) {
			continue
		}
		if err != nil {
			return domain.Settings{}, err
		}
		return domain.NewSettings(string(p), rc.URL, rc.Credential["token"])
	}
	return domain.Settings{}, modules.ErrNoConnection
}
