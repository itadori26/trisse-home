package adguard

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/adguard/domain"
	"waylight/internal/modules/adguard/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares AdGuard Home and Pi-hole as two services
// (#610). Pi-hole needs no user name; the check reads the statistics like
// SaveSettings.
func connectionSpecs(providers map[domain.Provider]usecase.DNSFilterProvider) []modules.ConnectionSpec {
	fields := map[domain.Provider][]modules.CredentialField{
		domain.ProviderAdGuard: {{Name: "username"}, {Name: "password", Secret: true}},
		domain.ProviderPiHole:  {{Name: "username", Optional: true}, {Name: "password", Secret: true}},
	}
	out := make([]modules.ConnectionSpec, 0, len(fields))
	for _, p := range []domain.Provider{domain.ProviderAdGuard, domain.ProviderPiHole} {
		provider := providers[p]
		out = append(out, modules.ConnectionSpec{
			Service: string(p), Title: p.Title(), Fields: fields[p],
			Check: func(ctx context.Context, url string, cred map[string]string) error {
				s, err := domain.NewSettings(p, url, cred["username"], cred["password"])
				if err != nil {
					return err
				}
				if _, err := provider.Stats(ctx, s); err != nil {
					return kernel.SetupFailure(p.Title(), err)
				}
				return nil
			},
		})
	}
	return out
}

// savedConnections returns one connection per configured DNS filter.
func savedConnections(ctx context.Context, repo usecase.SettingsRepository) ([]modules.SavedConnection, error) {
	var out []modules.SavedConnection
	for _, p := range []domain.Provider{domain.ProviderAdGuard, domain.ProviderPiHole} {
		s, err := repo.Load(ctx, p)
		if errors.Is(err, domain.ErrNotConfigured) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, modules.SavedConnection{Source: "adguard:" + string(p), Service: string(p), URL: s.URL,
			Credential: map[string]string{"username": s.Username, "password": s.Password}})
	}
	return out, nil
}

// connections adapts the integrations resolver to usecase.Connections; the
// connection's service is the provider.
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
func (c connections) Settings(ctx context.Context, p domain.Provider, id string) (domain.Settings, error) {
	rc, err := c.r.ResolveConnection(ctx, string(p), id)
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.NewSettings(p, rc.URL, rc.Credential["username"], rc.Credential["password"])
}
