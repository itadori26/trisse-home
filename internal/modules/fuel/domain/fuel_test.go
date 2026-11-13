package domain

import (
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

const key = "00000000-0000-0000-0000-000000000002"
const station = "51d4b660-a095-1aa0-e100-80009459e03a"

func TestNewSettings(t *testing.T) {
	s, err := NewSettings(" "+key+" ", 52.52, 13.4, " Berlin ", 5, "e10")
	if err != nil || s.Place != "Berlin" || s.APIKey != key {
		t.Fatalf("unexpected %+v %v", s, err)
	}
	bad := []struct {
		key           string
		lat, lon, rad float64
		fuel          string
	}{
		{"abc", 52, 13, 5, "e5"}, {key, 0, 0, 5, "e5"}, {key, 95, 13, 5, "e5"}, {key, 52, 13, 30, "e5"}, {key, 52, 13, 5, "lpg"},
	}
	for _, b := range bad {
		if _, err := NewSettings(b.key, b.lat, b.lon, "", b.rad, b.fuel); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%+v accepted", b)
		}
	}
}

func TestSetFavourites(t *testing.T) {
	var s Settings
	if err := s.SetFavourites([]Favourite{{ID: station, Name: " Aral "}, {ID: station, Name: "dup"}}); err != nil || len(s.Favourites) != 1 || s.Favourites[0].Name != "Aral" {
		t.Fatalf("duplicates dropped, names trimmed: %+v %v", s.Favourites, err)
	}
	if err := s.SetFavourites([]Favourite{{ID: "x", Name: "y"}}); err == nil {
		t.Fatal("invalid ID accepted")
	}
	many := make([]Favourite, MaxFavourites+1)
	for i := range many {
		many[i] = Favourite{ID: station, Name: "x"}
	}
	if err := s.SetFavourites(many); err == nil {
		t.Fatal("limit")
	}
}

func TestSortCheapest(t *testing.T) {
	list := []Station{
		{ID: "closed", Price: 1.5, Open: false},
		{ID: "far", Price: 1.7, Open: true, DistanceKm: 4},
		{ID: "none", Open: true},
		{ID: "near", Price: 1.7, Open: true, DistanceKm: 1},
		{ID: "cheap", Price: 1.6, Open: true, DistanceKm: 9},
	}
	SortCheapest(list)
	var got []string
	for _, s := range list {
		got = append(got, s.ID)
	}
	want := []string{"cheap", "near", "far", "closed", "none"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order %v, want %v", got, want)
		}
	}
}
