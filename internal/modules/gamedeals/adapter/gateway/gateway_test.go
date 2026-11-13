package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"trisse-home/internal/modules/gamedeals/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewSettingsRepository(dbtest.Open(t, "gamedeals", Migrations()))
	if s, err := repo.Load(ctx); err != nil || s != domain.DefaultSettings() {
		t.Fatalf("defaults: %+v %v", s, err)
	}
	s := domain.Settings{SteamID: "76561197960287930", Country: "AT", MinDiscount: 40, Epic: true}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.Load(ctx); err != nil || got != s {
		t.Fatalf("round trip: %+v %v", got, err)
	}
}

func TestSteam(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IWishlistService/GetWishlist/v1/":
			_, _ = w.Write([]byte(`{"response":{"items":[{"appid":620},{"appid":400}]}}`))
		case "/IStoreBrowseService/GetItems/v1/":
			var in struct {
				IDs     []struct{ AppID int } `json:"ids"`
				Context struct {
					CountryCode string `json:"country_code"`
				} `json:"context"`
			}
			_ = json.Unmarshal([]byte(r.URL.Query().Get("input_json")), &in)
			if len(in.IDs) != 2 || in.Context.CountryCode != "DE" {
				t.Errorf("input: %+v", in)
			}
			_, _ = w.Write([]byte(`{"response":{"store_items":[
				{"appid":620,"name":"Portal 2","best_purchase_option":{"formatted_final_price":"1,99€","final_price_in_cents":"199","original_price_in_cents":"999","discount_pct":80,"active_discounts":[{"discount_end_date":1794600000}]}},
				{"appid":400,"name":"Portal","best_purchase_option":{"formatted_final_price":"9,75€","final_price_in_cents":"975","discount_pct":0}}]}}`))
		}
	}))
	defer srv.Close()
	deals, err := NewSteam(srv.Client(), srv.URL).Wishlist(context.Background(), "76561197960287930", "DE")
	if err != nil || len(deals) != 1 || deals[0].Name != "Portal 2" || deals[0].FinalCents != 199 || deals[0].EndsAt != time.Unix(1794600000, 0).UTC() {
		t.Fatalf("only discounted items: %+v %v", deals, err)
	}
}

func TestEpic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("country") != "DE" {
			t.Errorf("query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"data":{"Catalog":{"searchStore":{"elements":[
			{"title":"Free Now","productSlug":"free-now","promotions":{"promotionalOffers":[{"promotionalOffers":[{"startDate":"2026-11-12T16:00:00.000Z","endDate":"2026-11-19T16:00:00.000Z","discountSetting":{"discountPercentage":0}}]}],"upcomingPromotionalOffers":[]}},
			{"title":"Free Next","catalogNs":{"mappings":[{"pageSlug":"free-next"}]},"promotions":{"promotionalOffers":[],"upcomingPromotionalOffers":[{"promotionalOffers":[{"startDate":"2026-11-19T16:00:00.000Z","endDate":"2026-11-26T16:00:00.000Z","discountSetting":{"discountPercentage":0}}]}]}},
			{"title":"Only Discounted","promotions":{"promotionalOffers":[{"promotionalOffers":[{"startDate":"2026-11-12T16:00:00.000Z","endDate":"2026-11-19T16:00:00.000Z","discountSetting":{"discountPercentage":50}}]}]}},
			{"title":"No promotion","promotions":null}]}}}}`))
	}))
	defer srv.Close()
	games, err := NewEpic(srv.Client(), srv.URL).FreeGames(context.Background(), "DE")
	if err != nil || len(games) != 2 || games[0].URL != "https://store.epicgames.com/p/free-now" || games[1].URL != "https://store.epicgames.com/p/free-next" {
		t.Fatalf("free games only: %+v %v", games, err)
	}
}
