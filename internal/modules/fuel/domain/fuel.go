package domain

import (
	"cmp"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// FuelType is a fuel sold at German stations.
type FuelType string

// Fuel types of the Markttransparenzstelle.
const (
	FuelE5     FuelType = "e5"
	FuelE10    FuelType = "e10"
	FuelDiesel FuelType = "diesel"
)

// Limits of the module.
const (
	MinRadiusKm   = 1.0
	MaxRadiusKm   = 25.0
	MaxFavourites = 10
	// MaxCheapest is how many stations the widget lists.
	MaxCheapest = 10
)

// ErrNotConfigured is returned while no API key and location are saved.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "Fuel prices are not configured yet.")

// stationIDPattern matches Tankerkönig station IDs (UUIDs).
var stationIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Favourite is a station the user follows, with the name shown when the
// station is outside the search radius.
type Favourite struct {
	ID, Name string
}

// Settings configure the search.
type Settings struct {
	APIKey   string
	Lat, Lon float64
	// Place is the label of the location, e.g. "Berlin".
	Place      string
	RadiusKm   float64
	Fuel       FuelType
	Favourites []Favourite
}

// NewSettings validates the settings.
func NewSettings(apiKey string, lat, lon float64, place string, radius float64, fuel string) (Settings, error) {
	s := Settings{APIKey: strings.TrimSpace(apiKey), Lat: lat, Lon: lon, Place: strings.TrimSpace(place), RadiusKm: radius, Fuel: FuelType(fuel)}
	if !stationIDPattern.MatchString(s.APIKey) {
		return Settings{}, kernel.NewValidationError("apiKey", "The API key must look like 00000000-0000-0000-0000-000000000002.")
	}
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 || (lat == 0 && lon == 0) {
		return Settings{}, kernel.NewValidationError("lat", "Please choose a location.")
	}
	if radius < MinRadiusKm || radius > MaxRadiusKm {
		return Settings{}, kernel.NewValidationError("radiusKm", "The radius must be between 1 and 25 km.")
	}
	if !slices.Contains([]FuelType{FuelE5, FuelE10, FuelDiesel}, s.Fuel) {
		return Settings{}, kernel.NewValidationError("fuel", "Choose E5, E10 or Diesel.")
	}
	if utf8.RuneCountInString(s.Place) > 80 {
		return Settings{}, kernel.NewValidationError("place", "The place name is too long.")
	}
	return s, nil
}

// SetFavourites validates and replaces the favourites.
func (s *Settings) SetFavourites(list []Favourite) error {
	if len(list) > MaxFavourites {
		return kernel.NewValidationErrorf("favourites", "At most {max} favourite stations.", "max", MaxFavourites)
	}
	out := make([]Favourite, 0, len(list))
	for _, f := range list {
		f.Name = strings.TrimSpace(f.Name)
		if !stationIDPattern.MatchString(f.ID) || f.Name == "" || utf8.RuneCountInString(f.Name) > 80 {
			return kernel.NewValidationError("favourites", "Every favourite needs a station ID and a name.")
		}
		if !slices.ContainsFunc(out, func(x Favourite) bool { return x.ID == f.ID }) {
			out = append(out, f)
		}
	}
	s.Favourites = out
	return nil
}

// Station is a fuel station with the price of the chosen fuel.
type Station struct {
	ID, Name, Brand, Address string
	DistanceKm               float64
	// Price is in euro; zero when the station reports none.
	Price float64
	Open  bool
}

// SortCheapest orders open stations with a price by price, then distance;
// closed stations follow, stations without a price come last.
func SortCheapest(list []Station) {
	slices.SortStableFunc(list, func(a, b Station) int {
		if ra, rb := rank(a), rank(b); ra != rb {
			return ra - rb
		}
		if c := cmp.Compare(a.Price, b.Price); c != 0 {
			return c
		}
		return cmp.Compare(a.DistanceKm, b.DistanceKm)
	})
}

func rank(s Station) int {
	switch {
	case s.Price <= 0:
		return 2
	case !s.Open:
		return 1
	}
	return 0
}

// Prices are the current prices of one station by fuel type.
type Prices struct {
	Open   bool
	ByFuel map[FuelType]float64
}

// Overview is what the widget shows.
type Overview struct {
	Fuel       FuelType
	Place      string
	Cheapest   []Station
	Favourites []Station
}
