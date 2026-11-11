// Package presenter maps parcels to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/modules/parcels/usecase"
)

// SettingsDTO is returned by the settings endpoints (never the key).
type SettingsDTO struct {
	HasDHLAPIKey      bool `json:"hasDhlApiKey"`
	HideDeliveredDays int  `json:"hideDeliveredDays"`
}

// ParcelDTO is one tracked parcel.
type ParcelDTO struct {
	ID                string     `json:"id"`
	TrackingNumber    string     `json:"trackingNumber"`
	Carrier           string     `json:"carrier" tstype:"'dhl'"`
	Label             string     `json:"label"`
	Status            string     `json:"status" tstype:"'unknown' | 'pre-transit' | 'transit' | 'delivered' | 'failure'"`
	Description       string     `json:"description"`
	Location          string     `json:"location"`
	StatusAt          *time.Time `json:"statusAt"`
	EstimatedDelivery *time.Time `json:"estimatedDelivery"`
	Error             string     `json:"error"`
}

func optional(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Parcel maps one parcel.
func Parcel(p domain.Parcel) ParcelDTO {
	return ParcelDTO{ID: p.ID, TrackingNumber: p.TrackingNumber, Carrier: string(p.Carrier), Label: p.Label,
		Status: string(p.Status.Code), Description: p.Status.Description, Location: p.Status.Location,
		StatusAt: optional(p.Status.At), EstimatedDelivery: optional(p.EstimatedDelivery), Error: p.Error}
}

// Parcels maps a list of parcels.
func Parcels(ps []domain.Parcel) []ParcelDTO {
	out := make([]ParcelDTO, 0, len(ps))
	for _, p := range ps {
		out = append(out, Parcel(p))
	}
	return out
}
