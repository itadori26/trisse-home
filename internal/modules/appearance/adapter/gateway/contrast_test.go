package gateway

import (
	"math"
	"strconv"
	"testing"
)

// luminance is the WCAG relative luminance of #rgb or #rrggbb.
func luminance(hex string) float64 {
	h := hex[1:]
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	var l float64
	for i, w := range []float64{0.2126, 0.7152, 0.0722} {
		v, _ := strconv.ParseUint(h[2*i:2*i+2], 16, 8)
		c := float64(v) / 255
		if c <= 0.04045 {
			c /= 12.92
		} else {
			c = math.Pow((c+0.055)/1.055, 2.4)
		}
		l += w * c
	}
	return l
}

func contrast(a, b string) float64 {
	la, lb := luminance(a), luminance(b)
	return (math.Max(la, lb) + 0.05) / (math.Min(la, lb) + 0.05)
}

// Every built-in theme meets WCAG AA (4.5:1) for all text it is used for (#113).
func TestBuiltinThemesMeetWCAGAA(t *testing.T) {
	pairs := map[string][]string{
		"--color-text":            {"--color-bg", "--color-surface", "--color-surface-2"},
		"--color-text-muted":      {"--color-bg", "--color-surface", "--color-surface-2"},
		"--color-accent":          {"--color-surface"},
		"--color-danger":          {"--color-surface"},
		"--color-accent-contrast": {"--color-accent"},
	}
	for _, th := range BuiltinThemes() {
		for fg, bgs := range pairs {
			for _, bg := range bgs {
				a, okA := th.Tokens[fg]
				b, okB := th.Tokens[bg]
				if !okA || !okB {
					continue
				}
				if c := contrast(a, b); c < 4.5 {
					t.Errorf("%s: %s %s on %s %s = %.2f:1, want 4.5:1", th.ID, fg, a, bg, b, c)
				}
			}
		}
	}
}
