package gateway

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"trisse-home/internal/modules/shopping/domain"
	"trisse-home/internal/shared/ical"
)

const maxResponseBytes = 4 << 20

const todoQuery = `<?xml version="1.0" encoding="utf-8"?>
<C:calendar-query xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop><D:getetag/><C:calendar-data/></D:prop>
  <C:filter><C:comp-filter name="VCALENDAR"><C:comp-filter name="VTODO"/></C:comp-filter></C:filter>
</C:calendar-query>`

type multistatus struct {
	Responses []struct {
		Href      string `xml:"href"`
		Propstats []struct {
			ETag         string `xml:"prop>getetag"`
			CalendarData string `xml:"prop>calendar-data"`
		} `xml:"propstat"`
	} `xml:"response"`
}

// CalDAVStore implements usecase.Store.
type CalDAVStore struct {
	client *http.Client
	now    func() time.Time
}

// NewCalDAVStore returns a store.
func NewCalDAVStore(client *http.Client) *CalDAVStore {
	return &CalDAVStore{client: client, now: time.Now}
}

func (c *CalDAVStore) do(ctx context.Context, s domain.Settings, method, target string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(s.Username, s.Password)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return res, nil
}

func statusError(res *http.Response) error {
	switch res.StatusCode {
	case http.StatusPreconditionFailed:
		return domain.ErrConflict
	case http.StatusNotFound, http.StatusGone:
		return domain.ErrItemNotFound
	default:
		return fmt.Errorf("server answered %s", res.Status)
	}
}

// List implements usecase.Store.
func (c *CalDAVStore) List(ctx context.Context, s domain.Settings) ([]domain.Item, error) {
	res, err := c.do(ctx, s, "REPORT", s.ListURL, strings.NewReader(todoQuery),
		map[string]string{"Depth": "1", "Content-Type": "application/xml; charset=utf-8"})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusMultiStatus {
		return nil, statusError(res)
	}
	var ms multistatus
	if err := xml.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(&ms); err != nil {
		return nil, fmt.Errorf("parse CalDAV response: %w", err)
	}
	var items []domain.Item
	for _, r := range ms.Responses {
		id := path.Base(r.Href)
		if !domain.ValidItemID(id) {
			continue
		}
		for _, ps := range r.Propstats {
			for _, t := range ical.ParseTodos(ps.CalendarData) {
				items = append(items, domain.Item{ID: id, Summary: t.Summary, Done: t.Done, Created: t.Created, Completed: t.Completed})
			}
		}
	}
	return items, nil
}

func objectURL(s domain.Settings, id string) string { return s.ListURL + id }

// Add implements usecase.Store; If-None-Match never overwrites an object.
func (c *CalDAVStore) Add(ctx context.Context, s domain.Settings, uid, summary string) (domain.Item, error) {
	id := url.PathEscape(strings.ReplaceAll(uid, "@", "-")) + ".ics"
	now := c.now()
	res, err := c.do(ctx, s, http.MethodPut, objectURL(s, id), strings.NewReader(ical.NewTodo(uid, summary, now)),
		map[string]string{"Content-Type": "text/calendar; charset=utf-8", "If-None-Match": "*"})
	if err != nil {
		return domain.Item{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		return domain.Item{}, statusError(res)
	}
	return domain.Item{ID: id, Summary: summary, Created: now}, nil
}

// SetDone implements usecase.Store: the object is read, only its completion
// properties are changed and it is written back with If-Match, so changes
// made meanwhile by other apps are never overwritten.
func (c *CalDAVStore) SetDone(ctx context.Context, s domain.Settings, id string, done bool) error {
	res, err := c.do(ctx, s, http.MethodGet, objectURL(s, id), nil, nil)
	if err != nil {
		return err
	}
	raw, readErr := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes))
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return statusError(res)
	}
	if readErr != nil {
		return fmt.Errorf("read item: %w", readErr)
	}
	etag := res.Header.Get("ETag")
	headers := map[string]string{"Content-Type": "text/calendar; charset=utf-8"}
	if etag != "" {
		headers["If-Match"] = etag
	}
	put, err := c.do(ctx, s, http.MethodPut, objectURL(s, id), bytes.NewBufferString(ical.SetTodoDone(string(raw), done, c.now())), headers)
	if err != nil {
		return err
	}
	defer put.Body.Close()
	if put.StatusCode != http.StatusNoContent && put.StatusCode != http.StatusCreated && put.StatusCode != http.StatusOK {
		return statusError(put)
	}
	return nil
}

// Delete implements usecase.Store.
func (c *CalDAVStore) Delete(ctx context.Context, s domain.Settings, id string) error {
	res, err := c.do(ctx, s, http.MethodDelete, objectURL(s, id), nil, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return statusError(res)
	}
	return nil
}
