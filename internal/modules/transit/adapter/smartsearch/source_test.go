package smartsearch

import (
	"context"
	"testing"
	"time"

	"waylight/internal/modules/transit/domain"
	"waylight/internal/shared/answers"
)

type fakeTransit struct {
	stops      []domain.Stop
	departures []domain.Departure
	asked      string
}

func (f *fakeTransit) SearchStops(_ context.Context, q string) ([]domain.Stop, error) {
	f.asked = q
	return f.stops, nil
}

func (f *fakeTransit) StopsNear(context.Context, float64, float64) ([]domain.Stop, error) {
	return f.stops, nil
}

func (f *fakeTransit) Departures(_ context.Context, stopID string, _ int) ([]domain.Departure, error) {
	f.asked += "|" + stopID
	return f.departures, nil
}

func TestTransitAnswer(t *testing.T) {
	now := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	tr := &fakeTransit{
		stops: []domain.Stop{{ID: "de:hbf", Name: "Oldenburg Hbf"}},
		departures: []domain.Departure{
			{Line: "S1", Headsign: "Airport", Planned: now.Add(4 * time.Minute), Expected: now.Add(10 * time.Minute), RealTime: true, Track: "3"},
			{Line: "7", Headsign: "Zoo", Planned: now.Add(5 * time.Minute), Cancelled: true},
		},
	}
	s := New(tr)
	named := answers.Match{Intent: intents[0], Slots: map[string]answers.SlotValue{"place": {Type: answers.SlotPlace, Text: "oldenburg hbf"}}}
	a, err := s.Resolve(context.Background(), answers.Request{Match: named, Lang: answers.German, Now: now})
	if err != nil || a.Title != "Oldenburg Hbf" || tr.asked != "oldenburg hbf|de:hbf" {
		t.Fatalf("answer = %+v, %v (asked %q)", a, err, tr.asked)
	}
	if a.Items[0].Value != "10 Min. · +6 · 3" || a.Items[0].State != answers.StateWarning || a.Items[1].Value != "fällt aus" {
		t.Errorf("items = %+v", a.Items)
	}

	// Without a name the stop of the transit widget is used.
	tr.asked = ""
	widget := func(_ context.Context, kind string) map[string]any {
		if kind == "transit" {
			return map[string]any{"stopId": "de:home", "stopName": "Home"}
		}
		return nil
	}
	a, _ = s.Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: intents[0]}, Lang: answers.English, Now: now, Widget: widget})
	if a.Title != "Home" || tr.asked != "|de:home" {
		t.Errorf("widget stop: %+v (asked %q)", a, tr.asked)
	}

	// Neither: ask for a stop.
	a, _ = s.Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: intents[0]}, Lang: answers.English, Now: now})
	if a.Kind != answers.Text || a.Title != "Departures" {
		t.Errorf("no stop: %+v", a)
	}
}
