package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/modules/parcels/usecase"
	"trisse-home/internal/shared/kernel"
)

// DefaultDHLURL is the DHL Shipment Tracking – Unified API.
const DefaultDHLURL = "https://api-eu.dhl.com/track/shipments"

// dhlMinInterval respects the free plan (one call per five seconds).
const dhlMinInterval = 5 * time.Second

const maxDHLBytes = 1 << 20

// DHL implements usecase.Provider for DHL.
type DHL struct {
	client   *http.Client
	endpoint string
	interval time.Duration

	mu   sync.Mutex
	last time.Time
}

// NewDHL returns a DHL provider.
func NewDHL(client *http.Client, endpoint string) *DHL {
	return &DHL{client: client, endpoint: endpoint, interval: dhlMinInterval}
}

// wait enforces the minimum interval between calls.
func (d *DHL) wait(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if wait := d.interval - time.Since(d.last); wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	d.last = time.Now()
	return nil
}

type dhlResponse struct {
	Shipments []struct {
		Status struct {
			Timestamp   string `json:"timestamp"`
			StatusCode  string `json:"statusCode"`
			Status      string `json:"status"`
			Description string `json:"description"`
			Location    struct {
				Address struct {
					Locality string `json:"addressLocality"`
				} `json:"address"`
			} `json:"location"`
		} `json:"status"`
		EstimatedTimeOfDelivery string `json:"estimatedTimeOfDelivery"`
		EstimatedDeliveryWindow struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"estimatedTimeOfDeliveryTimeFrame"`
	} `json:"shipments"`
}

// Track implements usecase.Provider.
func (d *DHL) Track(ctx context.Context, s domain.Settings, trackingNumber string) (domain.Status, usecase.Estimate, error) {
	if err := d.wait(ctx); err != nil {
		return domain.Status{}, usecase.Estimate{}, err
	}
	q := url.Values{"trackingNumber": {trackingNumber}, "language": {"de"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return domain.Status{}, usecase.Estimate{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("DHL-API-Key", s.DHLAPIKey)
	req.Header.Set("Accept", "application/json")
	res, err := d.client.Do(req)
	if err != nil {
		return domain.Status{}, usecase.Estimate{}, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized, http.StatusForbidden:
		return domain.Status{}, usecase.Estimate{}, kernel.NewPublicError(kernel.ErrUnauthorized, "DHL rejected the API key.")
	case http.StatusNotFound:
		return domain.Status{}, usecase.Estimate{}, kernel.NewPublicError(kernel.ErrNotFound, "DHL does not know this tracking number (yet).")
	case http.StatusTooManyRequests:
		return domain.Status{}, usecase.Estimate{}, kernel.NewPublicError(kernel.ErrTooManyRequests, "The daily DHL limit is reached; it will be retried later.")
	default:
		return domain.Status{}, usecase.Estimate{}, fmt.Errorf("DHL answered %s", res.Status)
	}
	var body dhlResponse
	if err := json.NewDecoder(io.LimitReader(res.Body, maxDHLBytes)).Decode(&body); err != nil {
		return domain.Status{}, usecase.Estimate{}, fmt.Errorf("decode response: %w", err)
	}
	if len(body.Shipments) == 0 {
		return domain.Status{Code: domain.StatusUnknown}, usecase.Estimate{}, nil
	}
	sh := body.Shipments[0]
	st := domain.Status{Code: statusCode(sh.Status.StatusCode), Location: strings.TrimSpace(sh.Status.Location.Address.Locality)}
	st.Description = strings.TrimSpace(sh.Status.Description)
	if st.Description == "" {
		st.Description = strings.TrimSpace(sh.Status.Status)
	}
	st.At = parseDHLTime(sh.Status.Timestamp)
	var est usecase.Estimate
	if t := parseDHLTime(sh.EstimatedDeliveryWindow.To); !t.IsZero() {
		est = usecase.Estimate{From: parseDHLTime(sh.EstimatedDeliveryWindow.From).Unix(), To: t.Unix()}
	} else if t := parseDHLTime(sh.EstimatedTimeOfDelivery); !t.IsZero() {
		est = usecase.Estimate{From: t.Unix(), To: t.Unix()}
	}
	return st, est, nil
}

func statusCode(s string) domain.StatusCode {
	switch domain.StatusCode(strings.ToLower(s)) {
	case domain.StatusPreTransit:
		return domain.StatusPreTransit
	case domain.StatusTransit:
		return domain.StatusTransit
	case domain.StatusDelivered:
		return domain.StatusDelivered
	case domain.StatusFailure:
		return domain.StatusFailure
	default:
		return domain.StatusUnknown
	}
}

// parseDHLTime accepts RFC 3339 with or without zone (DHL omits it at times).
func parseDHLTime(s string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
