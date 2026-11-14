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
	Carrier           string     `json:"carrier" tstype:"'dhl' | 'hermes' | 'dpd' | 'gls' | 'ups'"`
	Label             string     `json:"label"`
	Status            string     `json:"status" tstype:"'unknown' | 'pre-transit' | 'transit' | 'delivered' | 'failure'"`
	Description       string     `json:"description"`
	Location          string     `json:"location"`
	StatusAt          *time.Time `json:"statusAt"`
	EstimatedDelivery *time.Time `json:"estimatedDelivery"`
	Error             string     `json:"error"`
	// Discovered parcels were found in a shipping e-mail (#166).
	Discovered bool `json:"discovered"`
	// TrackingURL is the carrier's public tracking page.
	TrackingURL string `json:"trackingUrl"`
	// Trackable is false for carriers without a tracking API: only the link.
	Trackable bool `json:"trackable"`
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
		StatusAt: optional(p.Status.At), EstimatedDelivery: optional(p.EstimatedDelivery), Error: p.Error,
		Discovered: p.Discovered, TrackingURL: p.TrackingURL(), Trackable: p.Carrier == domain.CarrierDHL}
}

// Parcels maps a list of parcels.
func Parcels(ps []domain.Parcel) []ParcelDTO {
	out := make([]ParcelDTO, 0, len(ps))
	for _, p := range ps {
		out = append(out, Parcel(p))
	}
	return out
}

// MailSettingsDTO is returned by the mail endpoints (#166); never the password.
type MailSettingsDTO struct {
	Enabled         bool       `json:"enabled"`
	Host            string     `json:"host"`
	Port            int        `json:"port"`
	Username        string     `json:"username"`
	HasPassword     bool       `json:"hasPassword"`
	Folder          string     `json:"folder"`
	IntervalMinutes int        `json:"intervalMinutes"`
	LastRun         *time.Time `json:"lastRun"`
	LastError       string     `json:"lastError"`
	Found           int        `json:"found"`
}

// MailSettings maps a MailSettingsView.
func MailSettings(v usecase.MailSettingsView) MailSettingsDTO {
	return MailSettingsDTO{Enabled: v.Enabled, Host: v.Host, Port: v.Port, Username: v.Username, HasPassword: v.HasPassword,
		Folder: v.Folder, IntervalMinutes: v.IntervalMinutes, LastRun: optional(v.LastRun), LastError: v.LastError, Found: v.Found}
}
