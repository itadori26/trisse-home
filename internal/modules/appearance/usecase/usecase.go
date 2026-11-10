// Package usecase contains the appearance use cases.
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/appearance/domain"
)

// AppearanceResponse is the use case output.
type AppearanceResponse struct {
	Theme           string
	AccentColor     string
	BackgroundImage string
	CustomCSS       string
}

// UpdateAppearanceRequest is the input of UpdateAppearance.
type UpdateAppearanceRequest struct {
	Theme           string
	AccentColor     string
	BackgroundImage string
	CustomCSS       string
}

// GetAppearance returns the current preferences (input port).
type GetAppearance interface {
	Execute(ctx context.Context) (AppearanceResponse, error)
}

// UpdateAppearance validates and stores new preferences (input port).
type UpdateAppearance interface {
	Execute(ctx context.Context, req UpdateAppearanceRequest) (AppearanceResponse, error)
}

// Repository persists the appearance (output port). Load returns
// domain.Default() when nothing has been stored yet.
type Repository interface {
	Load(ctx context.Context) (domain.Appearance, error)
	Save(ctx context.Context, a domain.Appearance) error
}

func toResponse(a domain.Appearance) AppearanceResponse {
	return AppearanceResponse{Theme: string(a.Theme()), AccentColor: a.AccentColor(), BackgroundImage: a.BackgroundImage(), CustomCSS: a.CustomCSS()}
}

type getAppearance struct{ repo Repository }

// NewGetAppearance returns the GetAppearance use case.
func NewGetAppearance(repo Repository) GetAppearance { return &getAppearance{repo: repo} }

// Execute implements GetAppearance.
func (uc *getAppearance) Execute(ctx context.Context) (AppearanceResponse, error) {
	a, err := uc.repo.Load(ctx)
	if err != nil {
		return AppearanceResponse{}, fmt.Errorf("load appearance: %w", err)
	}
	return toResponse(a), nil
}

type updateAppearance struct{ repo Repository }

// NewUpdateAppearance returns the UpdateAppearance use case.
func NewUpdateAppearance(repo Repository) UpdateAppearance { return &updateAppearance{repo: repo} }

// Execute implements UpdateAppearance.
func (uc *updateAppearance) Execute(ctx context.Context, req UpdateAppearanceRequest) (AppearanceResponse, error) {
	a, err := domain.NewAppearance(domain.Theme(req.Theme), req.AccentColor, req.BackgroundImage, req.CustomCSS)
	if err != nil {
		return AppearanceResponse{}, err
	}
	if err := uc.repo.Save(ctx, a); err != nil {
		return AppearanceResponse{}, fmt.Errorf("save appearance: %w", err)
	}
	return toResponse(a), nil
}
