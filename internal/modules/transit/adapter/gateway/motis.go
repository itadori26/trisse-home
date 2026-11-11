// Package gateway queries a MOTIS journey planner API (e.g. Transitous).
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/transit/domain"
)

// DefaultURL is the public, community-run Transitous API.
const DefaultURL = "https://api.transitous.org"

const (
	maxResponseBytes = 2 << 20
	maxStops         = 8
	userAgent        = "trisse-home (self-hosted start page)"
)

// MOTIS implements usecase.Provider.
type MOTIS struct {
	client *http.Client
	base   string
}

// NewMOTIS returns a client for baseURL.
func NewMOTIS(client *http.Client, baseURL string) *MOTIS {
	return &MOTIS{client: client, base: strings.TrimRight(baseURL, "/")}
}

func (m *MOTIS) get(ctx context.Context, path string, q url.Values, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.base+path+"?"+q.Encode(), nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	res, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("transit API answered %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// SearchStops implements usecase.Provider.
func (m *MOTIS) SearchStops(ctx context.Context, query string) ([]domain.Stop, error) {
	var matches []struct {
		ID   string  `json:"id"`
		Name string  `json:"name"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
	}
	if err := m.get(ctx, "/api/v1/geocode", url.Values{"text": {query}, "type": {"STOP"}}, &matches); err != nil {
		return nil, err
	}
	out := make([]domain.Stop, 0, min(len(matches), maxStops))
	for _, s := range matches {
		if domain.ValidStopID(s.ID) {
			out = append(out, domain.Stop{ID: s.ID, Name: s.Name, Lat: s.Lat, Lon: s.Lon})
		}
		if len(out) == maxStops {
			break
		}
	}
	return out, nil
}

// Departures implements usecase.Provider.
func (m *MOTIS) Departures(ctx context.Context, stopID string, n int) ([]domain.Departure, error) {
	var res struct {
		StopTimes []struct {
			Place struct {
				Departure          string `json:"departure"`
				ScheduledDeparture string `json:"scheduledDeparture"`
				Track              string `json:"track"`
			} `json:"place"`
			Mode           string `json:"mode"`
			RealTime       bool   `json:"realTime"`
			Headsign       string `json:"headsign"`
			RouteShortName string `json:"routeShortName"`
			Cancelled      bool   `json:"cancelled"`
		} `json:"stopTimes"`
	}
	q := url.Values{"stopId": {stopID}, "n": {strconv.Itoa(n)}}
	if err := m.get(ctx, "/api/v1/stoptimes", q, &res); err != nil {
		return nil, err
	}
	out := make([]domain.Departure, 0, len(res.StopTimes))
	for _, st := range res.StopTimes {
		planned, err := time.Parse(time.RFC3339, st.Place.ScheduledDeparture)
		if err != nil {
			continue // arrivals at the terminus have no departure
		}
		d := domain.Departure{Line: st.RouteShortName, Headsign: st.Headsign, Mode: strings.ToLower(st.Mode),
			Planned: planned, RealTime: st.RealTime, Track: st.Place.Track, Cancelled: st.Cancelled}
		d.Expected, _ = time.Parse(time.RFC3339, st.Place.Departure)
		out = append(out, d)
	}
	return out, nil
}
