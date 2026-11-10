package config

import (
	"strings"
	"testing"
	"time"
)

const validKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=" // 32 bytes

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddr != ":8090" || cfg.CookieSecure || cfg.SessionTTL != 30*24*time.Hour {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if len(cfg.AllowedCIDRs) != 6 {
		t.Fatalf("expected 6 default CIDRs, got %d", len(cfg.AllowedCIDRs))
	}
	if !cfg.ModuleEnabled("weather") {
		t.Fatal("weather should be enabled by default")
	}
}

func TestLoadCollectsAllErrors(t *testing.T) {
	_, err := Load(env(map[string]string{
		"ALLOWED_CIDRS": "nope",
		"COOKIE_SECURE": "maybe",
		"SESSION_TTL":   "-1h",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	for _, key := range []string{"ALLOWED_CIDRS", "COOKIE_SECURE", "SESSION_TTL", "SECRET_KEY"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error does not mention %s: %v", key, err)
		}
	}
}

func TestDisabledModules(t *testing.T) {
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey, "MODULES_DISABLED": "Docker, adguard"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ModuleEnabled("docker") || cfg.ModuleEnabled("adguard") || !cfg.ModuleEnabled("rss") {
		t.Fatalf("unexpected module switches: %v", cfg.DisabledModules)
	}
}
