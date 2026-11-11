package domain

import (
	"errors"
	"testing"
	"time"
)

func TestDownsampleLatency(t *testing.T) {
	var ps []LatencyPoint
	for i := range 6 {
		ps = append(ps, LatencyPoint{At: time.Unix(int64(i), 0), LatencyMs: float64(i), Online: i != 4})
	}
	got := DownsampleLatency(ps, 3)
	if len(got) != 3 || got[0].LatencyMs != 0.5 || got[2].Online || !got[1].Online {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestStatusFromCheckAndCertificate(t *testing.T) {
	now := time.Unix(0, 0)
	s := StatusFromCheck("1", CheckResult{HTTPStatus: 200, CertExpiresAt: now.Add(10 * 24 * time.Hour)}, now)
	if s.State != StateOnline || !s.CertificateExpiresWithin(DefaultCertWarning, now) || s.CertificateExpiresWithin(24*time.Hour, now) {
		t.Fatalf("unexpected %+v", s)
	}
	if StatusFromCheck("1", CheckResult{Err: errors.New("x")}, now).CertificateExpiresWithin(DefaultCertWarning, now) {
		t.Fatal("missing certificate must not be flagged")
	}
}
