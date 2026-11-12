package domain

import (
	"errors"
	"testing"
	"time"
)

func TestPasskeyCounter(t *testing.T) {
	now := time.Unix(0, 0)
	p, err := NewPasskey("1", OwnerID, []byte("c"), []byte("k"), -7, 5, "", now)
	if err != nil || p.Name != "Passkey" {
		t.Fatalf("%+v %v", p, err)
	}
	if err := p.Use(6, now); err != nil {
		t.Fatal(err)
	}
	for _, c := range []uint32{6, 3, 0} {
		if err := p.Use(c, now); !errors.Is(err, ErrPasskeyCloned) {
			t.Errorf("counter %d accepted after 6", c)
		}
	}
	zero, _ := NewPasskey("2", OwnerID, []byte("c"), []byte("k"), -8, 0, "Key", now)
	for range 2 { // every use reports 0
		if err := zero.Use(0, now); err != nil {
			t.Fatal("authenticators without counters must work")
		}
	}
}
