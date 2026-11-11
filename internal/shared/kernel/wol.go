package kernel

import (
	"bytes"
	"encoding/hex"
	"strings"
)

// MagicPacketSize is the length of a Wake-on-LAN magic packet: 6 × 0xFF
// followed by the target MAC address repeated 16 times.
const MagicPacketSize = 6 + 16*6

// MAC is a normalised hardware address ("aa:bb:cc:dd:ee:ff") or empty.
type MAC string

// ParseMAC accepts aa:bb:cc:dd:ee:ff, aa-bb-cc-dd-ee-ff, aabb.ccdd.eeff and
// aabbccddeeff (any case). An empty string yields an empty MAC.
func ParseMAC(raw string) (MAC, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	hexDigits := strings.NewReplacer(":", "", "-", "", ".", "").Replace(strings.ToLower(raw))
	b, err := hex.DecodeString(hexDigits)
	if err != nil || len(b) != 6 {
		return "", NewValidationError("mac", "The MAC address must look like aa:bb:cc:dd:ee:ff.")
	}
	parts := make([]string, 6)
	for i, x := range b {
		parts[i] = hex.EncodeToString([]byte{x})
	}
	return MAC(strings.Join(parts, ":")), nil
}

func (m MAC) bytes() []byte {
	b, _ := hex.DecodeString(strings.ReplaceAll(string(m), ":", "")) // normalised by ParseMAC
	return b
}

// MagicPacket builds the Wake-on-LAN packet for m.
func MagicPacket(m MAC) []byte {
	p := bytes.Repeat([]byte{0xFF}, 6)
	return append(p, bytes.Repeat(m.bytes(), 16)...)
}

// ParseMagicPacket validates a magic packet and returns its target MAC.
func ParseMagicPacket(p []byte) (MAC, bool) {
	if len(p) != MagicPacketSize || !bytes.Equal(p[:6], bytes.Repeat([]byte{0xFF}, 6)) {
		return "", false
	}
	mac := p[6:12]
	if !bytes.Equal(p[6:], bytes.Repeat(mac, 16)) {
		return "", false
	}
	parts := make([]string, 6)
	for i, x := range mac {
		parts[i] = hex.EncodeToString([]byte{x})
	}
	return MAC(strings.Join(parts, ":")), true
}
