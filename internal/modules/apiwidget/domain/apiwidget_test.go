package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func okPath(string) error { return nil }

func f(v float64) *float64 { return &v }

func valid() EndpointInput {
	return EndpointInput{Name: " NAS ", URL: "http://192.168.1.5:8080/api/status?x=1", Extractions: []Extraction{{Label: "Temp", Path: "$.temp", Unit: "°C"}}}
}

func TestNewEndpoint(t *testing.T) {
	e, err := NewEndpoint("a", valid(), nil, okPath)
	if err != nil || e.Name != "NAS" || e.Interval != DefaultInterval || e.URL != "http://192.168.1.5:8080/api/status?x=1" {
		t.Fatalf("%+v %v", e, err)
	}
	for name, mod := range map[string]func(*EndpointInput){
		"no name":        func(i *EndpointInput) { i.Name = " " },
		"bad url":        func(i *EndpointInput) { i.URL = "ftp://x" },
		"metadata":       func(i *EndpointInput) { i.URL = "http://169.254.169.254/latest" },
		"credentials":    func(i *EndpointInput) { i.URL = "http://u:p@nas/api" },
		"short interval": func(i *EndpointInput) { i.Interval = 5 * time.Second },
		"long interval":  func(i *EndpointInput) { i.Interval = 48 * time.Hour },
		"no values":      func(i *EndpointInput) { i.Extractions = nil },
		"too many":       func(i *EndpointInput) { i.Extractions = make([]Extraction, 11) },
		"no label":       func(i *EndpointInput) { i.Extractions[0].Label = "" },
		"no path":        func(i *EndpointInput) { i.Extractions[0].Path = " " },
		"long unit":      func(i *EndpointInput) { i.Extractions[0].Unit = strings.Repeat("x", 13) },
		"header no val":  func(i *EndpointInput) { i.HeaderName = "X-Api-Key" },
		"bad header":     func(i *EndpointInput) { i.HeaderName, i.HeaderValue = "X Key", "v" },
		"host header":    func(i *EndpointInput) { i.HeaderName, i.HeaderValue = "Host", "evil" },
		"crlf value":     func(i *EndpointInput) { i.HeaderName, i.HeaderValue = "X-Key", "a\r\nX-Other: b" },
	} {
		in := valid()
		mod(&in)
		if _, err := NewEndpoint("a", in, nil, okPath); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := NewEndpoint("a", valid(), nil, func(string) error { return errors.New("bad") }); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("path check: %v", err)
	}
}

func TestHeaderSecretIsKept(t *testing.T) {
	stored := Endpoint{HeaderName: "X-Api-Key", HeaderValue: "secret"}
	in := valid()
	in.HeaderName, in.KeepHeader = "x-api-key", true
	e, err := NewEndpoint("a", in, &stored, okPath)
	if err != nil || e.HeaderValue != "secret" {
		t.Fatalf("keep: %+v %v", e, err)
	}
	// A different header name does not inherit the secret.
	in.HeaderName = "Authorization"
	if _, err := NewEndpoint("a", in, &stored, okPath); err == nil {
		t.Fatal("secret must not move to another header")
	}
	// No header name removes the header.
	in.HeaderName = ""
	if e, _ := NewEndpoint("a", in, &stored, okPath); e.HasHeader() || e.HeaderValue != "" {
		t.Fatalf("remove: %+v", e)
	}
}

func TestEvaluate(t *testing.T) {
	above := Extraction{Label: "CPU", Unit: "%", Warn: f(80), Crit: f(95)}
	below := Extraction{Label: "Free", Warn: f(20), Crit: f(5), Below: true}
	for _, c := range []struct {
		x      Extraction
		v      any
		status Status
		text   string
	}{
		{above, 50.0, StatusOK, "50"},
		{above, 80.0, StatusWarning, "80"},
		{above, 99.5, StatusCritical, "99.5"},
		{above, RawNumber("96"), StatusCritical, "96"},
		{above, " 81 ", StatusWarning, " 81 "},
		{below, 30.0, StatusOK, "30"},
		{below, 20.0, StatusWarning, "20"},
		{below, 1.0, StatusCritical, "1"},
		{Extraction{}, 3.0, StatusOK, "3"},
		{above, "online", StatusUnknown, "online"},
		{above, true, StatusUnknown, "true"},
		{above, nil, StatusUnknown, "null"},
		{above, Composite(`{"a":1}`), StatusUnknown, `{"a":1}`},
	} {
		got := c.x.Evaluate(c.v)
		if got.Status != c.status || got.Text != c.text {
			t.Errorf("%v: %+v, want %s %q", c.v, got, c.status, c.text)
		}
	}
	if v := above.Evaluate(strings.Repeat("x", 200)); len([]rune(v.Text)) != 80 {
		t.Fatalf("long text not cut: %d", len([]rune(v.Text)))
	}
	if m := above.Missing("not found"); m.Status != StatusUnknown || m.Error != "not found" || m.Label != "CPU" {
		t.Fatalf("%+v", m)
	}
}
