package kernel

import "testing"

func TestValidClientID(t *testing.T) {
	for id, want := range map[string]bool{
		"0b7e9d2c-4f1a-4c8e-9a51-2d3c4b5a6f70": true,
		"abcdefgh":                             true,
		"short":                                false,
		"../../etc":                            false,
		"has space here":                       false,
		"x" + string(make([]byte, 64)):         false,
	} {
		if got := ValidClientID(id); got != want {
			t.Errorf("%q: %v", id, got)
		}
	}
}
