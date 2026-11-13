package domain

import (
	"cmp"
	"regexp"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// steamIDPattern matches 64-bit Steam IDs (profile URLs show them as
// .../profiles/7656119…).
var steamIDPattern = regexp.MustCompile(`^7656119\d{10}$`)

var countryPattern = regexp.MustCompile(`^[A-Z]{2}$`)

// Settings configure the sources (#62). Nothing is requested while neither
// a Steam ID is set nor Epic is enabled.
type Settings struct {
	// SteamID is the 64-bit ID of a profile with a public wishlist; empty
	// disables Steam.
	SteamID string
	// Country selects prices and currency, e.g. "DE".
	Country     string
	MinDiscount int
	Epic        bool
}

// DefaultSettings are used until the user saves settings: everything off.
func DefaultSettings() Settings { return Settings{Country: "DE", MinDiscount: 20} }

// Enabled reports whether any source is configured.
func (s Settings) Enabled() bool { return s.SteamID != "" || s.Epic }

// NewSettings validates the settings.
func NewSettings(steamID, country string, minDiscount int, epic bool) (Settings, error) {
	s := Settings{SteamID: strings.TrimSpace(steamID), Country: strings.ToUpper(strings.TrimSpace(country)), MinDiscount: minDiscount, Epic: epic}
	if s.SteamID != "" && !steamIDPattern.MatchString(s.SteamID) {
		return Settings{}, kernel.NewValidationError("steamId", "The Steam ID must be the 17-digit number from your profile URL.")
	}
	if !countryPattern.MatchString(s.Country) {
		return Settings{}, kernel.NewValidationError("country", "The country must be a two-letter code such as DE.")
	}
	if minDiscount < 0 || minDiscount > 100 {
		return Settings{}, kernel.NewValidationError("minDiscount", "The minimum discount must be between 0 and 100 %.")
	}
	return s, nil
}

// Deal is a discounted game from the wishlist.
type Deal struct {
	AppID         int
	Name          string
	DiscountPct   int
	FinalPrice    string // formatted by the store, e.g. "14,99€"
	EndsAt        time.Time
	OriginalCents int
	FinalCents    int
}

// FilterDeals keeps deals with at least min percent off, biggest discount first.
func FilterDeals(list []Deal, minPct int) []Deal {
	out := slices.DeleteFunc(slices.Clone(list), func(d Deal) bool { return d.DiscountPct <= 0 || d.DiscountPct < minPct })
	slices.SortStableFunc(out, func(a, b Deal) int {
		if c := cmp.Compare(b.DiscountPct, a.DiscountPct); c != 0 {
			return c
		}
		return strings.Compare(a.Name, b.Name)
	})
	return out
}

// FreeGame is a game that is (or will be) free on the Epic Games Store.
type FreeGame struct {
	Title      string
	URL        string
	Start, End time.Time
}

// Current reports whether the offer runs at now.
func (g FreeGame) Current(now time.Time) bool { return !now.Before(g.Start) && now.Before(g.End) }

// Overview is what the widget shows.
type Overview struct {
	Deals    []Deal
	FreeNow  []FreeGame
	FreeNext []FreeGame
	// Errors names the sources that could not be loaded.
	Errors []string
}
