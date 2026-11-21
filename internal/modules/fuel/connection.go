package fuel

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/fuel/domain"
	"waylight/internal/modules/fuel/usecase"
	"waylight/internal/shared/kernel"
)

// checkLat and checkLon are a point the check asks prices around:
// Tankerkönig has no request that tests only the key, and the place of the
// widget is a setting of the module, not of the connection.
const checkLat, checkLon = 52.52, 13.405

// connectionSpecs declares the Tankerkönig connection (#610); the address
// is fixed, only the API key is entered.
func connectionSpecs(provider usecase.FuelPriceProvider, apiURL string) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "tankerkoenig", Title: "Tankerkönig", FixedURL: apiURL,
		Fields: []modules.CredentialField{{Name: "apiKey", Secret: true}},
		Check: func(ctx context.Context, _ string, cred map[string]string) error {
			s, err := domain.NewSettings(cred["apiKey"], checkLat, checkLon, "", domain.MinRadiusKm, string(domain.FuelE5))
			if err != nil {
				return err
			}
			if _, err := provider.Around(ctx, s); err != nil {
				return kernel.SetupFailure("Tankerkönig", err)
			}
			return nil
		},
	}}
}

// savedConnections returns the API key of the module settings as a
// connection to copy; place, radius and favourites stay in the module.
func savedConnections(ctx context.Context, repo usecase.SettingsRepository, apiURL string) ([]modules.SavedConnection, error) {
	s, err := repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []modules.SavedConnection{{Source: "fuel", Service: "tankerkoenig", URL: apiURL, Credential: map[string]string{"apiKey": s.APIKey}}}, nil
}

// connections adapts the integrations resolver to usecase.Connections. The
// address of a Tankerkönig connection is fixed, so only the key is used.
type connections struct{ r modules.ConnectionResolver }

// APIKey implements usecase.Connections.
func (c connections) APIKey(ctx context.Context, id string) (string, error) {
	rc, err := c.r.ResolveConnection(ctx, "tankerkoenig", id)
	if err != nil {
		return "", err
	}
	return rc.Credential["apiKey"], nil
}
