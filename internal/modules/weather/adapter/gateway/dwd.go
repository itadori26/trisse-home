package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"trisse-home/internal/modules/weather/domain"
)

// Default DWD open data endpoints (GeoServer of the Deutscher Wetterdienst).
const (
	DefaultDWDWFS = "https://maps.dwd.de/geoserver/dwd/ows"
	DefaultDWDWMS = "https://maps.dwd.de/geoserver/dwd/wms"
	radarLayers   = "dwd:Laender,dwd:Niederschlagsradar"
	maxDWDBytes   = 4 << 20
)

// DWD implements usecase.WarningProvider and usecase.RadarProvider.
type DWD struct {
	client   *http.Client
	wfs, wms string
}

// NewDWD returns a client for the given WFS and WMS endpoints.
func NewDWD(client *http.Client, wfsURL, wmsURL string) *DWD {
	return &DWD{client: client, wfs: wfsURL, wms: wmsURL}
}

func (d *DWD) get(ctx context.Context, u string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, "", fmt.Errorf("build request: %w", err)
	}
	res, err := d.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("DWD answered %s", res.Status)
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, maxDWDBytes))
	return b, res.Header.Get("Content-Type"), err
}

type warningFeatures struct {
	Features []struct {
		Properties struct {
			Identifier  string `json:"IDENTIFIER"`
			Event       string `json:"EVENT"`
			Headline    string `json:"HEADLINE"`
			Description string `json:"DESCRIPTION"`
			Instruction string `json:"INSTRUCTION"`
			Severity    string `json:"SEVERITY"`
			Onset       string `json:"ONSET"`
			Expires     string `json:"EXPIRES"`
		} `json:"properties"`
	} `json:"features"`
}

// Warnings implements usecase.WarningProvider with a WFS point query on the
// municipality warnings (axis order of EPSG:4326 in WFS 2.0 is lat, lon).
func (d *DWD) Warnings(ctx context.Context, loc domain.Location) ([]domain.Warning, error) {
	q := url.Values{
		"service": {"WFS"}, "version": {"2.0.0"}, "request": {"GetFeature"},
		"typeName": {"dwd:Warnungen_Gemeinden"}, "outputFormat": {"application/json"},
		"propertyName": {"IDENTIFIER,EVENT,HEADLINE,DESCRIPTION,INSTRUCTION,SEVERITY,ONSET,EXPIRES"},
		"CQL_FILTER":   {fmt.Sprintf("CONTAINS(THE_GEOM,POINT(%s %s))", coord(loc.Lat()), coord(loc.Lon()))},
	}
	body, _, err := d.get(ctx, d.wfs+"?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var fc warningFeatures
	if err := json.Unmarshal(body, &fc); err != nil {
		return nil, fmt.Errorf("decode warnings: %w", err)
	}
	out := make([]domain.Warning, 0, len(fc.Features))
	for _, f := range fc.Features {
		p := f.Properties
		w := domain.Warning{ID: p.Identifier, Event: p.Event, Headline: p.Headline, Description: p.Description,
			Instruction: p.Instruction, Severity: domain.ParseSeverity(p.Severity)}
		w.Onset, _ = time.Parse(time.RFC3339, p.Onset)
		w.Expires, _ = time.Parse(time.RFC3339, p.Expires)
		out = append(out, w)
	}
	return out, nil
}

func coord(v float64) string { return strconv.FormatFloat(v, 'f', 4, 64) }

var pngSignature = []byte("\x89PNG\r\n\x1a\n")

// Radar implements usecase.RadarProvider with a WMS GetMap request.
func (d *DWD) Radar(ctx context.Context, box domain.BBox, width, height int) ([]byte, error) {
	q := url.Values{
		"service": {"WMS"}, "version": {"1.3.0"}, "request": {"GetMap"}, "layers": {radarLayers}, "styles": {""},
		"format": {"image/png"}, "transparent": {"true"}, "crs": {"EPSG:4326"},
		"bbox":  {fmt.Sprintf("%s,%s,%s,%s", coord(box.MinLat), coord(box.MinLon), coord(box.MaxLat), coord(box.MaxLon))},
		"width": {strconv.Itoa(width)}, "height": {strconv.Itoa(height)},
	}
	body, _, err := d.get(ctx, d.wms+"?"+q.Encode())
	if err != nil {
		return nil, err
	}
	// GeoServer reports errors as XML with status 200.
	if !bytes.HasPrefix(body, pngSignature) {
		return nil, errors.New("DWD returned no image")
	}
	return body, nil
}
