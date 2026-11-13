package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"trisse-home/internal/modules/fuel/domain"
)

// DefaultTankerkoenigURL is the public API (data: Markttransparenzstelle
// für Kraftstoffe, licence CC BY 4.0).
const DefaultTankerkoenigURL = "https://creativecommons.tankerkoenig.de/json"

const (
	maxResponseBytes = 1 << 20
	// maxPriceIDs is the provider's limit for prices.php.
	maxPriceIDs = 10
)

// Tankerkoenig implements usecase.FuelPriceProvider.
type Tankerkoenig struct {
	client  *http.Client
	baseURL string
}

// NewTankerkoenig returns a client for baseURL (DefaultTankerkoenigURL).
func NewTankerkoenig(client *http.Client, baseURL string) *Tankerkoenig {
	return &Tankerkoenig{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

type listPayload struct {
	OK       bool   `json:"ok"`
	Message  string `json:"message"`
	Stations []struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Brand       string   `json:"brand"`
		Street      string   `json:"street"`
		HouseNumber string   `json:"houseNumber"`
		Place       string   `json:"place"`
		Dist        float64  `json:"dist"`
		Price       *float64 `json:"price"`
		IsOpen      bool     `json:"isOpen"`
	} `json:"stations"`
}

// Around implements usecase.FuelPriceProvider.
func (t *Tankerkoenig) Around(ctx context.Context, s domain.Settings) ([]domain.Station, error) {
	q := url.Values{
		"lat": {strconv.FormatFloat(s.Lat, 'f', 5, 64)}, "lng": {strconv.FormatFloat(s.Lon, 'f', 5, 64)},
		"rad": {strconv.FormatFloat(s.RadiusKm, 'f', 1, 64)}, "sort": {"price"}, "type": {string(s.Fuel)}, "apikey": {s.APIKey},
	}
	var p listPayload
	if err := t.get(ctx, "/list.php?"+q.Encode(), &p); err != nil {
		return nil, err
	}
	if !p.OK {
		return nil, errors.New(p.Message)
	}
	out := make([]domain.Station, 0, len(p.Stations))
	for _, st := range p.Stations {
		station := domain.Station{ID: st.ID, Name: st.Name, Brand: st.Brand, DistanceKm: st.Dist, Open: st.IsOpen,
			Address: strings.TrimSpace(strings.TrimSpace(st.Street+" "+st.HouseNumber) + ", " + st.Place)}
		if st.Price != nil {
			station.Price = *st.Price
		}
		out = append(out, station)
	}
	return out, nil
}

// Prices implements usecase.FuelPriceProvider.
func (t *Tankerkoenig) Prices(ctx context.Context, apiKey string, ids []string) (map[string]domain.Prices, error) {
	if len(ids) > maxPriceIDs {
		ids = ids[:maxPriceIDs]
	}
	q := url.Values{"ids": {strings.Join(ids, ",")}, "apikey": {apiKey}}
	var p struct {
		OK      bool   `json:"ok"`
		Message string `json:"message"`
		Prices  map[string]struct {
			Status string          `json:"status"`
			E5     json.RawMessage `json:"e5"`
			E10    json.RawMessage `json:"e10"`
			Diesel json.RawMessage `json:"diesel"`
		} `json:"prices"`
	}
	if err := t.get(ctx, "/prices.php?"+q.Encode(), &p); err != nil {
		return nil, err
	}
	if !p.OK {
		return nil, errors.New(p.Message)
	}
	out := make(map[string]domain.Prices, len(p.Prices))
	for id, v := range p.Prices {
		out[id] = domain.Prices{Open: v.Status == "open", ByFuel: map[domain.FuelType]float64{
			domain.FuelE5: price(v.E5), domain.FuelE10: price(v.E10), domain.FuelDiesel: price(v.Diesel)}}
	}
	return out, nil
}

// price reads a price that is false when a station does not sell the fuel.
func price(raw json.RawMessage) float64 {
	var f float64
	if json.Unmarshal(raw, &f) != nil {
		return 0
	}
	return f
}

func (t *Tankerkoenig) get(ctx context.Context, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	res, err := t.client.Do(req)
	if err != nil {
		// The URL carries the API key; never put it into an error message.
		var uerr *url.Error
		if errors.As(err, &uerr) {
			return fmt.Errorf("request failed: %w", uerr.Err)
		}
		return errors.New("request failed")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected answer from Tankerkönig: %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(dst); err != nil {
		return fmt.Errorf("unexpected response from Tankerkönig: %w", err)
	}
	return nil
}
