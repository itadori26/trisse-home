package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"trisse-home/internal/modules/apiwidget/domain"
)

// JSONPath implements usecase.PathEvaluator with a small, dependency-free
// subset of JSONPath (ADR 0010):
//
//	$                  the document
//	.name              member (letters, digits, _ and -)
//	['any name']       member with any characters ("…" works as well)
//	[2] / [-1]         array element, negative counts from the end
//	.length()          number of elements of an array or object, or characters of a string
//
// The leading $ may be left out ("data.temp" is "$.data.temp").
// Filters and wildcards are not supported: the widget shows single values.
type JSONPath struct{}

type step struct {
	name   string
	index  int
	kind   int // stepName, stepIndex, stepLength
	quoted bool
}

const (
	stepName = iota
	stepIndex
	stepLength
)

// Validate implements usecase.PathEvaluator.
func (JSONPath) Validate(path string) error {
	_, err := compile(path)
	return err
}

// Evaluate implements usecase.PathEvaluator. The result is converted to the
// domain's value types: json.Number becomes domain.RawNumber, objects and
// arrays become domain.Composite.
func (p JSONPath) Evaluate(doc any, path string) (any, error) {
	v, err := p.find(doc, path)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case json.Number:
		return domain.RawNumber(t), nil
	case map[string]any, []any:
		raw, err := json.Marshal(t)
		if err != nil {
			return nil, fmt.Errorf("encode value: %w", err)
		}
		return domain.Composite(raw), nil
	}
	return v, nil
}

func (JSONPath) find(doc any, path string) (any, error) {
	steps, err := compile(path)
	if err != nil {
		return nil, err
	}
	cur := doc
	for _, s := range steps {
		switch s.kind {
		case stepLength:
			switch v := cur.(type) {
			case []any:
				return float64(len(v)), nil
			case map[string]any:
				return float64(len(v)), nil
			case string:
				return float64(len([]rune(v))), nil
			default:
				return nil, errors.New("length() needs an array, object or string")
			}
		case stepName:
			obj, ok := cur.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%q: not an object", s.name)
			}
			if cur, ok = obj[s.name]; !ok {
				return nil, fmt.Errorf("%q not found", s.name)
			}
		case stepIndex:
			arr, ok := cur.([]any)
			if !ok {
				return nil, fmt.Errorf("[%d]: not an array", s.index)
			}
			i := s.index
			if i < 0 {
				i += len(arr)
			}
			if i < 0 || i >= len(arr) {
				return nil, fmt.Errorf("[%d] out of range (%d elements)", s.index, len(arr))
			}
			cur = arr[i]
		}
	}
	return cur, nil
}

func compile(path string) ([]step, error) {
	p := strings.TrimSpace(path)
	switch {
	case strings.HasPrefix(p, "$"):
		p = p[1:]
	case strings.HasPrefix(p, ".") || strings.HasPrefix(p, "["):
	default:
		p = "." + p
	}
	var steps []step
	for p != "" {
		if len(steps) > 0 && steps[len(steps)-1].kind == stepLength {
			return nil, errors.New("length() must come last")
		}
		switch p[0] {
		case '.':
			p = p[1:]
			if strings.HasPrefix(p, "length()") {
				steps = append(steps, step{kind: stepLength})
				p = p[len("length()"):]
				continue
			}
			n := 0
			for n < len(p) && isNameByte(p[n]) {
				n++
			}
			if n == 0 {
				return nil, errors.New("expected a member name after '.'")
			}
			steps = append(steps, step{kind: stepName, name: p[:n]})
			p = p[n:]
		case '[':
			end := strings.IndexByte(p, ']')
			if end < 0 {
				return nil, errors.New("missing ']'")
			}
			inner := p[1:end]
			if len(inner) >= 2 && (inner[0] == '\'' || inner[0] == '"') {
				name, rest, err := quoted(p[1:])
				if err != nil {
					return nil, err
				}
				if !strings.HasPrefix(rest, "]") {
					return nil, errors.New("expected ']' after the quoted name")
				}
				steps = append(steps, step{kind: stepName, name: name, quoted: true})
				p = rest[1:]
				continue
			}
			i, err := strconv.Atoi(strings.TrimSpace(inner))
			if err != nil {
				return nil, fmt.Errorf("[%s]: only numbers or quoted names are supported", inner)
			}
			steps = append(steps, step{kind: stepIndex, index: i})
			p = p[end+1:]
		default:
			return nil, fmt.Errorf("unexpected %q", p[0])
		}
	}
	return steps, nil
}

// quoted reads 'name' or "name" (backslash escapes the next character) at
// the start of s and returns the name and the rest after the closing quote.
func quoted(s string) (string, string, error) {
	q := s[0]
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\\' && i+1 < len(s):
			i++
			b.WriteByte(s[i])
		case c == q:
			return b.String(), s[i+1:], nil
		default:
			b.WriteByte(c)
		}
	}
	return "", "", errors.New("missing closing quote")
}

func isNameByte(c byte) bool {
	return c == '_' || c == '-' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c >= 0x80
}
