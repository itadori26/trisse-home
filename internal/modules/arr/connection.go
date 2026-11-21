package arr

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/arr/domain"
	"waylight/internal/modules/arr/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares Sonarr, Radarr and Prowlarr as services of their
// own (#610); the check is the one AddInstance makes.
func connectionSpecs(providers map[domain.App]usecase.Provider) []modules.ConnectionSpec {
	out := make([]modules.ConnectionSpec, 0, len(providers))
	for _, app := range []domain.App{domain.AppSonarr, domain.AppRadarr, domain.AppProwlarr} {
		provider := providers[app]
		out = append(out, modules.ConnectionSpec{
			Service: string(app), Title: app.Title(),
			Fields: []modules.CredentialField{{Name: "apiKey", Secret: true}},
			Check: func(ctx context.Context, url string, cred map[string]string) error {
				in, err := domain.NewInstance("check", string(app), "", url, cred["apiKey"])
				if err != nil {
					return err
				}
				if err := provider.Check(ctx, in); err != nil {
					return kernel.SetupFailure(app.Title(), err)
				}
				return nil
			},
		})
	}
	return out
}

// savedConnections returns every instance as a connection of its app,
// named like the instance.
func savedConnections(ctx context.Context, repo usecase.InstanceRepository) ([]modules.SavedConnection, error) {
	list, err := repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]modules.SavedConnection, 0, len(list))
	for _, in := range list {
		out = append(out, modules.SavedConnection{Source: "arr:" + in.ID, Service: string(in.App), Name: in.Name, URL: in.URL,
			Credential: map[string]string{"apiKey": in.APIKey}})
	}
	return out, nil
}

// connections adapts the integrations resolver to usecase.Connections; a
// connection may be of any of the three apps.
type connections struct{ r modules.ConnectionResolver }

// Instance implements usecase.Connections; the connection's ID and name
// stand in for the instance's.
func (c connections) Instance(ctx context.Context, id string) (domain.Instance, error) {
	for _, app := range []domain.App{domain.AppSonarr, domain.AppRadarr, domain.AppProwlarr} {
		rc, err := c.r.ResolveConnection(ctx, string(app), id)
		if errors.Is(err, modules.ErrNoConnection) {
			continue
		}
		if err != nil {
			return domain.Instance{}, err
		}
		return domain.NewInstance(id, string(app), rc.Name, rc.URL, rc.Credential["apiKey"])
	}
	return domain.Instance{}, modules.ErrNoConnection
}
