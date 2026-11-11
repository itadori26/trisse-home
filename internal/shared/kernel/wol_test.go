package kernel

import "testing"

func TestMAC(t *testing.T) {
	for _, in := range []string{"AA:bb:CC:dd:EE:ff", "aa-bb-cc-dd-ee-ff", "aabb.ccdd.eeff", "aabbccddeeff"} {
		if m, err := ParseMAC(in); err != nil || m != "aa:bb:cc:dd:ee:ff" {
			t.Errorf("%s: %q %v", in, m, err)
		}
	}
	for _, in := range []string{"aa:bb:cc", "zz:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:ff:00"} {
		if _, err := ParseMAC(in); err == nil {
			t.Errorf("%s accepted", in)
		}
	}
	if m, err := ParseMAC(" "); err != nil || m != "" {
		t.Fatal("empty MAC must be allowed")
	}
}

func TestMagicPacket(t *testing.T) {
	p := MagicPacket("01:02:03:04:05:06")
	if len(p) != MagicPacketSize || p[0] != 0xFF || p[6] != 1 || p[101] != 6 {
		t.Fatalf("packet %x", p)
	}
	if m, ok := ParseMagicPacket(p); !ok || m != "01:02:03:04:05:06" {
		t.Fatalf("round trip %q %v", m, ok)
	}
	p[50] ^= 1
	if _, ok := ParseMagicPacket(p); ok {
		t.Fatal("corrupted packet accepted")
	}
	if _, ok := ParseMagicPacket([]byte("hello")); ok {
		t.Fatal("garbage accepted")
	}
}
