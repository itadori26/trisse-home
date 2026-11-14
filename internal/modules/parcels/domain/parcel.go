package domain

import (
	"cmp"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits and defaults.
const (
	MaxParcels               = 30
	DefaultHideDeliveredDays = 3
	maxHideDeliveredDays     = 30
	maxLabelLength           = 60
)

// Carrier identifies a tracking provider.
type Carrier string

// Supported carriers. Only DHL has a tracking provider; parcels of the
// others link to the carrier's tracking page (#166).
const (
	CarrierDHL    Carrier = "dhl"
	CarrierHermes Carrier = "hermes"
	CarrierDPD    Carrier = "dpd"
	CarrierGLS    Carrier = "gls"
	CarrierUPS    Carrier = "ups"
)

// Carriers lists the supported carriers.
var Carriers = []Carrier{CarrierDHL, CarrierHermes, CarrierDPD, CarrierGLS, CarrierUPS}

// trackingPages are the public tracking pages; %s is the number.
var trackingPages = map[Carrier]string{
	CarrierDHL:    "https://www.dhl.de/de/privatkunden/pakete-empfangen/verfolgen.html?piececode=%s",
	CarrierHermes: "https://www.myhermes.de/empfangen/sendungsverfolgung/sendungsinformation#%s",
	CarrierDPD:    "https://tracking.dpd.de/status/de_DE/parcel/%s",
	CarrierGLS:    "https://www.gls-pakete.de/sendungsverfolgung?match=%s",
	CarrierUPS:    "https://www.ups.com/track?tracknum=%s",
}

// TrackingURL is the carrier's public tracking page for the parcel.
func (p Parcel) TrackingURL() string {
	if page, ok := trackingPages[p.Carrier]; ok {
		return fmt.Sprintf(page, url.QueryEscape(p.TrackingNumber))
	}
	return ""
}

// StatusCode is the normalised shipment state.
type StatusCode string

// Status codes (aligned with the DHL unified API).
const (
	StatusUnknown    StatusCode = "unknown"
	StatusPreTransit StatusCode = "pre-transit"
	StatusTransit    StatusCode = "transit"
	StatusDelivered  StatusCode = "delivered"
	StatusFailure    StatusCode = "failure"
)

// Errors.
var (
	ErrParcelNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The parcel does not exist.")
	ErrTooMany        = kernel.NewPublicError(kernel.ErrConflict, "At most 30 parcels can be tracked.")
	ErrNoAPIKey       = kernel.NewPublicError(kernel.ErrInvalid, "Please add your DHL API key in the settings first.")
)

var trackingPattern = regexp.MustCompile(`^[A-Z0-9]{8,35}$`)

// NormaliseTrackingNumber removes spaces and validates the number.
func NormaliseTrackingNumber(s string) (string, error) {
	s = strings.ToUpper(strings.Join(strings.Fields(s), ""))
	if !trackingPattern.MatchString(s) {
		return "", kernel.NewValidationError("trackingNumber", "Tracking numbers consist of 8 to 35 letters and digits.")
	}
	return s, nil
}

// Settings configure tracking.
type Settings struct {
	DHLAPIKey         string
	HideDeliveredDays int
}

// NewSettings validates the settings.
func NewSettings(dhlKey string, hideDays int) (Settings, error) {
	if hideDays < 0 || hideDays > maxHideDeliveredDays {
		return Settings{}, kernel.NewValidationError("hideDeliveredDays", "Delivered parcels can be kept for 0 to 30 days.")
	}
	return Settings{DHLAPIKey: strings.TrimSpace(dhlKey), HideDeliveredDays: hideDays}, nil
}

// Status is the latest tracking event.
type Status struct {
	Code        StatusCode
	Description string
	Location    string
	At          time.Time
}

// Parcel is a tracked shipment.
type Parcel struct {
	ID                string
	TrackingNumber    string
	Carrier           Carrier
	Label             string
	Status            Status
	EstimatedDelivery time.Time
	CheckedAt         time.Time
	Error             string
	CreatedAt         time.Time
	// Discovered parcels were found in a shipping e-mail (#166).
	Discovered bool
}

// NewParcel validates a new parcel.
func NewParcel(id, trackingNumber string, carrier Carrier, label string, now time.Time) (Parcel, error) {
	tn, err := NormaliseTrackingNumber(trackingNumber)
	if err != nil {
		return Parcel{}, err
	}
	if !slices.Contains(Carriers, carrier) {
		return Parcel{}, kernel.NewValidationError("carrier", "This carrier is not supported.")
	}
	label = strings.TrimSpace(label)
	if utf8.RuneCountInString(label) > maxLabelLength {
		return Parcel{}, kernel.NewValidationError("label", "The label must be at most 60 characters long.")
	}
	return Parcel{ID: id, TrackingNumber: tn, Carrier: carrier, Label: label, CreatedAt: now, Status: Status{Code: StatusUnknown}}, nil
}

// Delivered reports whether the parcel arrived.
func (p Parcel) Delivered() bool { return p.Status.Code == StatusDelivered }

// Visible hides parcels delivered more than hideDays ago.
func (p Parcel) Visible(now time.Time, hideDays int) bool {
	return !p.Delivered() || now.Sub(p.Status.At) < time.Duration(hideDays)*24*time.Hour
}

// SortForDisplay shows undelivered parcels first (oldest first), then the
// most recent deliveries.
func SortForDisplay(ps []Parcel) {
	slices.SortStableFunc(ps, func(a, b Parcel) int {
		if a.Delivered() != b.Delivered() {
			if a.Delivered() {
				return 1
			}
			return -1
		}
		if a.Delivered() {
			return b.Status.At.Compare(a.Status.At)
		}
		return cmp.Compare(a.CreatedAt.Unix(), b.CreatedAt.Unix())
	})
}
