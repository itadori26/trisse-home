package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"trisse-home/internal/modules/gamedeals/domain"
)

// Public store endpoints. They are not official APIs: everything that
// depends on their shape lives in this file, and a change only empties one
// source of the widget (the other keeps working).
const (
	DefaultSteamAPIURL = "https://api.steampowered.com"
	DefaultEpicURL     = "https://store-site-backend-static.ak.epicgames.com"
	maxResponseBytes   = 4 << 20
	// maxWishlistItems bounds one GetItems request.
	maxWishlistItems = 200
)

// Steam implements usecase.WishlistProvider: IWishlistService lists the
// app IDs of a public wishlist, IStoreBrowseService adds names, prices and
// discount end dates in one request.
type Steam struct {
	client  *http.Client
	baseURL string
}

// NewSteam returns a client.
func NewSteam(client *http.Client, baseURL string) *Steam {
	return &Steam{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

// Wishlist implements usecase.WishlistProvider.
func (s *Steam) Wishlist(ctx context.Context, steamID, country string) ([]domain.Deal, error) {
	var list struct {
		Response struct {
			Items []struct {
				AppID int `json:"appid"`
			} `json:"items"`
		} `json:"response"`
	}
	if err := getJSON(ctx, s.client, s.baseURL+"/IWishlistService/GetWishlist/v1/?steamid="+url.QueryEscape(steamID), &list); err != nil {
		return nil, err
	}
	if len(list.Response.Items) == 0 {
		return []domain.Deal{}, nil
	}
	type id struct {
		AppID int `json:"appid"`
	}
	input := struct {
		IDs     []id `json:"ids"`
		Context struct {
			Language    string `json:"language"`
			CountryCode string `json:"country_code"`
		} `json:"context"`
		DataRequest struct {
			IncludeBasicInfo bool `json:"include_basic_info"`
		} `json:"data_request"`
	}{}
	for i, it := range list.Response.Items {
		if i == maxWishlistItems {
			break
		}
		input.IDs = append(input.IDs, id{it.AppID})
	}
	input.Context.Language, input.Context.CountryCode = "english", country
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	var items struct {
		Response struct {
			StoreItems []struct {
				AppID   int    `json:"appid"`
				Name    string `json:"name"`
				Success int    `json:"success"`
				Best    *struct {
					Final       string `json:"formatted_final_price"`
					FinalCents  string `json:"final_price_in_cents"`
					OrigCents   string `json:"original_price_in_cents"`
					DiscountPct int    `json:"discount_pct"`
					Active      []struct {
						EndDate int64 `json:"discount_end_date"`
					} `json:"active_discounts"`
				} `json:"best_purchase_option"`
			} `json:"store_items"`
		} `json:"response"`
	}
	if err := getJSON(ctx, s.client, s.baseURL+"/IStoreBrowseService/GetItems/v1/?input_json="+url.QueryEscape(string(raw)), &items); err != nil {
		return nil, err
	}
	out := []domain.Deal{}
	for _, it := range items.Response.StoreItems {
		if it.Best == nil || it.Best.DiscountPct <= 0 {
			continue
		}
		d := domain.Deal{AppID: it.AppID, Name: it.Name, DiscountPct: it.Best.DiscountPct, FinalPrice: it.Best.Final}
		_, _ = fmt.Sscan(it.Best.FinalCents, &d.FinalCents)
		_, _ = fmt.Sscan(it.Best.OrigCents, &d.OriginalCents)
		for _, a := range it.Best.Active {
			if a.EndDate > 0 {
				d.EndsAt = time.Unix(a.EndDate, 0).UTC()
				break
			}
		}
		out = append(out, d)
	}
	return out, nil
}

// Epic implements usecase.FreeGamesProvider with the store's promotions feed.
type Epic struct {
	client  *http.Client
	baseURL string
}

// NewEpic returns a client.
func NewEpic(client *http.Client, baseURL string) *Epic {
	return &Epic{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

type epicPromotion struct {
	PromotionalOffers []struct {
		StartDate       time.Time `json:"startDate"`
		EndDate         time.Time `json:"endDate"`
		DiscountSetting struct {
			DiscountPercentage int `json:"discountPercentage"`
		} `json:"discountSetting"`
	} `json:"promotionalOffers"`
}

// FreeGames implements usecase.FreeGamesProvider: offers with a 100 %
// discount (discountPercentage 0 means "free"), current and upcoming.
func (e *Epic) FreeGames(ctx context.Context, country string) ([]domain.FreeGame, error) {
	var p struct {
		Data struct {
			Catalog struct {
				SearchStore struct {
					Elements []struct {
						Title       string `json:"title"`
						ProductSlug string `json:"productSlug"`
						CatalogNs   struct {
							Mappings []struct {
								PageSlug string `json:"pageSlug"`
							} `json:"mappings"`
						} `json:"catalogNs"`
						Promotions *struct {
							Current  []epicPromotion `json:"promotionalOffers"`
							Upcoming []epicPromotion `json:"upcomingPromotionalOffers"`
						} `json:"promotions"`
					} `json:"elements"`
				} `json:"searchStore"`
			} `json:"Catalog"`
		} `json:"data"`
	}
	q := url.Values{"locale": {"en-US"}, "country": {country}, "allowCountries": {country}}
	if err := getJSON(ctx, e.client, e.baseURL+"/freeGamesPromotions?"+q.Encode(), &p); err != nil {
		return nil, err
	}
	var out []domain.FreeGame
	for _, el := range p.Data.Catalog.SearchStore.Elements {
		if el.Promotions == nil {
			continue
		}
		slug := el.ProductSlug
		if len(el.CatalogNs.Mappings) > 0 && el.CatalogNs.Mappings[0].PageSlug != "" {
			slug = el.CatalogNs.Mappings[0].PageSlug
		}
		link := ""
		if slug != "" {
			link = "https://store.epicgames.com/p/" + url.PathEscape(strings.TrimSuffix(slug, "/home"))
		}
		for _, group := range append(el.Promotions.Current, el.Promotions.Upcoming...) {
			for _, o := range group.PromotionalOffers {
				if o.DiscountSetting.DiscountPercentage == 0 {
					out = append(out, domain.FreeGame{Title: el.Title, URL: link, Start: o.StartDate, End: o.EndDate})
				}
			}
		}
	}
	return out, nil
}

func getJSON(ctx context.Context, client *http.Client, u string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("the store answered %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(dst); err != nil {
		return fmt.Errorf("unexpected response from the store: %w", err)
	}
	return nil
}
