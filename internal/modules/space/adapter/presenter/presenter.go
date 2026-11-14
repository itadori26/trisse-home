// Package presenter maps launches and the picture of the day to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/space/domain"
)

// LaunchDTO is one upcoming launch.
type LaunchDTO struct {
	ID          string    `json:"id"`
	Mission     string    `json:"mission"`
	MissionType string    `json:"missionType"`
	Description string    `json:"description"`
	Rocket      string    `json:"rocket"`
	Provider    string    `json:"provider"`
	Pad         string    `json:"pad"`
	Location    string    `json:"location"`
	Net         time.Time `json:"net"`
	// Status is go, tbd, tbc, hold, in_flight, success, failure, partial_failure or unknown.
	Status      string `json:"status"`
	WebcastLive bool   `json:"webcastLive"`
}

// PictureDTO is the picture of the day; Image is the proxied image URL or
// empty on video days, Link the original page or video.
type PictureDTO struct {
	Date        string `json:"date"`
	Title       string `json:"title"`
	Explanation string `json:"explanation"`
	Copyright   string `json:"copyright"`
	MediaType   string `json:"mediaType"`
	Image       string `json:"image"`
	Link        string `json:"link"`
}

// SettingsDTO is returned by the settings endpoints; the key is never returned.
type SettingsDTO struct {
	Launches bool `json:"launches"`
	Picture  bool `json:"picture"`
	HasKey   bool `json:"hasKey"`
}

// ImagePath is the stable URL of today's proxied picture (usable as a background).
const ImagePath = "/api/space/apod/image"

// Launches maps launches.
func Launches(ls []domain.Launch) []LaunchDTO {
	out := make([]LaunchDTO, 0, len(ls))
	for _, l := range ls {
		out = append(out, LaunchDTO{ID: l.ID, Mission: l.Mission, MissionType: l.MissionType, Description: l.Description, Rocket: l.Rocket,
			Provider: l.Provider, Pad: l.Pad, Location: l.Location, Net: l.Net, Status: string(l.Status), WebcastLive: l.WebcastLive})
	}
	return out
}

// Picture maps the picture of the day.
func Picture(p domain.Picture) PictureDTO {
	dto := PictureDTO{Date: p.Date, Title: p.Title, Explanation: p.Explanation, Copyright: p.Copyright, MediaType: p.MediaType, Link: p.URL}
	if p.IsImage() {
		dto.Image = ImagePath + "?d=" + p.Date
		dto.Link = "https://apod.nasa.gov/apod/astropix.html"
	}
	return dto
}

// Settings maps the settings.
func Settings(s domain.Settings) SettingsDTO {
	return SettingsDTO{Launches: s.Launches, Picture: s.Picture, HasKey: s.HasKey()}
}
