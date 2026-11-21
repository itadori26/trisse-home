package proxmox

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/proxmox/domain"
	"waylight/internal/modules/proxmox/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Proxmox VE connection (#610). The token ID
// and the pinned certificate fingerprint are not secret; the check reads
// the cluster like SaveSettings.
func connectionSpecs(client usecase.Client) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "proxmox", Title: "Proxmox VE",
		Fields: []modules.CredentialField{{Name: "tokenId"}, {Name: "secret", Secret: true}, {Name: "fingerprint", Optional: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["tokenId"], cred["secret"], cred["fingerprint"])
			if err != nil {
				return err
			}
			if _, err := client.Cluster(ctx, s); err != nil {
				return kernel.SetupFailure("Proxmox VE", err)
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
	return []modules.SavedConnection{{Source: "proxmox", Service: "proxmox", URL: s.URL,
		Credential: map[string]string{"tokenId": s.TokenID, "secret": s.Secret, "fingerprint": s.Fingerprint}}}, nil
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
	rc, err := c.r.ResolveConnection(ctx, "proxmox", id)
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.NewSettings(rc.URL, rc.Credential["tokenId"], rc.Credential["secret"], rc.Credential["fingerprint"])
}
