package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"trisse-home/internal/modules/apiwidget/domain"
	"trisse-home/internal/modules/apiwidget/usecase"
	"trisse-home/internal/shared/kernel"
)

// MaxResponseBytes is the largest JSON response that is read (1 MB).
const MaxResponseBytes = 1 << 20

// previewBytes is how much of the body the settings preview shows.
const previewBytes = 4 << 10

// requestTimeout bounds one request.
const requestTimeout = 10 * time.Second

// Fetcher implements usecase.JSONFetcher. Private is used unless the endpoint
// allows public addresses; both clients check every connection
// (httpclient.WithIntegrationGuard, WithPrivateOnly), so neither DNS nor a
// redirect can get around the address rule.
type Fetcher struct {
	Private *http.Client
	Public  *http.Client
}

var (
	errNotPrivate = kernel.NewPublicError(kernel.ErrInvalid,
		"The address is not in a private network. Allow public addresses for this endpoint to reach it.")
	errTooLarge = kernel.NewPublicError(kernel.ErrInvalid, "The response is larger than 1 MB.")
	errNotJSON  = kernel.NewPublicError(kernel.ErrInvalid, "The response is not JSON.")
	errTimeout  = kernel.NewPublicError(kernel.ErrInvalid, "The endpoint did not answer within 10 seconds.")
)

// Fetch implements usecase.JSONFetcher.
func (f Fetcher) Fetch(ctx context.Context, e domain.Endpoint) (usecase.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.URL, nil)
	if err != nil {
		return usecase.Response{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if e.HeaderName != "" {
		req.Header.Set(e.HeaderName, e.HeaderValue)
	}
	client := f.Private
	if e.AllowPublic {
		client = f.Public
	}
	res, err := client.Do(req)
	if err != nil {
		switch {
		case errors.Is(err, kernel.ErrTargetBlocked) && !e.AllowPublic:
			return usecase.Response{}, errNotPrivate
		case errors.Is(err, context.DeadlineExceeded):
			return usecase.Response{}, errTimeout
		}
		return usecase.Response{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, MaxResponseBytes+1))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return usecase.Response{}, errTimeout
		}
		return usecase.Response{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if len(raw) > MaxResponseBytes {
		return usecase.Response{}, errTooLarge
	}
	out := usecase.Response{Raw: raw[:min(len(raw), previewBytes)]}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return out, kernel.NewPublicErrorf(kernel.ErrInvalid, "The endpoint answered {status}.", "status", res.Status)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&out.Doc); err != nil {
		return out, errNotJSON
	}
	return out, nil
}
