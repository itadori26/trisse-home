package speedtest

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/speedtest/domain"
	"waylight/internal/modules/speedtest/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Speedtest Tracker connection (#610). The
// thresholds stay settings of the module; the check reads the results like
// SaveSettings.
func connectionSpecs(source usecase.SpeedtestSource, clock kernel.Clock) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "speedtest", Title: "Speedtest Tracker",
		Fields: []modules.CredentialField{{Name: "token", Secret: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["token"], 0, 0)
			if err != nil {
				return err
			}
			if _, err := source.Results(ctx, s, clock.Now().Add(-usecase.HistoryRange)); err != nil {
				return kernel.SetupFailure("Speedtest Tracker", err)
			}
			return nil
		},
	}}
}

// savedConnections returns the module settings as a connection to copy;
// the thresholds stay in the module.
func savedConnections(ctx context.Context, repo usecase.SettingsRepository) ([]modules.SavedConnection, error) {
	s, err := repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []modules.SavedConnection{{Source: "speedtest", Service: "speedtest", URL: s.URL, Credential: map[string]string{"token": s.Token}}}, nil
}

// connections adapts the integrations resolver to usecase.Connections.
type connections struct{ r modules.ConnectionResolver }

// Settings implements usecase.Connections; the thresholds stay in the
// module settings.
func (c connections) Settings(ctx context.Context, id string) (domain.Settings, error) {
	rc, err := c.r.ResolveConnection(ctx, "speedtest", id)
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.NewSettings(rc.URL, rc.Credential["token"], 0, 0)
}
