package gateway

import (
	"encoding/json"
	"strings"
	"testing"

	"trisse-home/internal/modules/apiwidget/domain"
)

func TestJSONPath(t *testing.T) {
	var doc any
	_ = json.Unmarshal([]byte(`{"data":{"temp":21.5,"rooms":[{"name":"Bad"},{"name":"Küche","on":true}],"a b":"x","ü-key":3},
		"list":[1,2,3],"s":"häh"}`), &doc)
	p := JSONPath{}
	for path, want := range map[string]any{
		"$.data.temp":           21.5,
		"data.temp":             21.5,
		"$.data.rooms[1].name":  "Küche",
		"$.data.rooms[-1].on":   true,
		"$['data']['a b']":      "x",
		`$["data"].ü-key`:       3.0,
		"$.list.length()":       3.0,
		"$.data.rooms.length()": 2.0,
		"$.s.length()":          3.0,
		"$.data.length()":       4.0,
		`$['data']["a b"]`:      "x",
		"$.data['quote\\'s']":   nil, // compiles; evaluated below as missing
	} {
		got, err := p.Evaluate(doc, path)
		if want == nil {
			if err == nil {
				t.Errorf("%s: expected not found", path)
			}
			continue
		}
		if err != nil || got != want {
			t.Errorf("%s = %v (%v), want %v", path, got, err, want)
		}
	}
	for _, bad := range []string{"$.", "$[", "$[x]", "$['open]", "$.list.length().x", "$..deep", "$[*]", "$.a?b"} {
		if p.Validate(bad) == nil {
			t.Errorf("%q must be invalid", bad)
		}
	}
	for _, path := range []string{"$.missing", "$.list[3]", "$.list.x", "$.data.temp[0]", "$.data.temp.length()"} {
		if _, err := p.Evaluate(doc, path); err == nil {
			t.Errorf("%s: expected an error", path)
		}
	}
}

func TestJSONPathConvertsToDomainValues(t *testing.T) {
	var doc any
	dec := json.NewDecoder(strings.NewReader(`{"list":[1,2],"n":12345678901234567890,"o":{"a":true}}`))
	dec.UseNumber()
	_ = dec.Decode(&doc)
	for path, want := range map[string]any{
		"$.list": domain.Composite("[1,2]"),
		"$.o":    domain.Composite(`{"a":true}`),
		"$.n":    domain.RawNumber("12345678901234567890"),
		"$.o.a":  true,
	} {
		if got, err := (JSONPath{}).Evaluate(doc, path); err != nil || got != want {
			t.Errorf("%s = %#v (%v), want %#v", path, got, err, want)
		}
	}
}
