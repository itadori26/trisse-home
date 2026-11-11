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

func TestDiskPathsKeepCase(t *testing.T) {
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey, "SYSTEM_DISK_PATHS": "Root=/host/rootfs, Data=/mnt/Data"}))
	if err != nil || cfg.DiskPaths[1] != "Data=/mnt/Data" {
		t.Fatalf("unexpected %v %v", cfg.DiskPaths, err)
	}
}

func TestSampleSettings(t *testing.T) {
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey, "SAMPLE_WEATHER_LOCATION": "Oldenburg, 53.14, 8.21"}))
	if err != nil || !cfg.SampleData || cfg.SampleWeather != (SampleLocation{Name: "Oldenburg", Lat: 53.14, Lon: 8.21}) || len(cfg.SampleFeeds) != 2 {
		t.Fatalf("unexpected %+v %v", cfg, err)
	}
	if _, err := Load(env(map[string]string{"SECRET_KEY": validKey, "SAMPLE_WEATHER_LOCATION": "Nowhere,100,0"})); err == nil {
		t.Fatal("expected error for invalid coordinates")
	}
}

func TestBackupDefaults(t *testing.T) {
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey, "DATABASE_PATH": "/data/trisse-home.db"}))
	if err != nil || cfg.BackupDir != "/data/backups" || cfg.BackupInterval != 24*time.Hour || cfg.BackupRetention != 7 {
		t.Fatalf("unexpected %q %v %d %v", cfg.BackupDir, cfg.BackupInterval, cfg.BackupRetention, err)
	}
	if _, err := Load(env(map[string]string{"SECRET_KEY": validKey, "BACKUP_RETENTION": "many"})); err == nil {
		t.Fatal("expected error for non-numeric retention")
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

func TestDockerActionsNeedAllowList(t *testing.T) {
	if _, err := Load(env(map[string]string{"SECRET_KEY": validKey, "DOCKER_ACTIONS": "true"})); err == nil || !strings.Contains(err.Error(), "DOCKER_ACTIONS_CONTAINERS") {
		t.Fatalf("expected allow-list error, got %v", err)
	}
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey, "DOCKER_ACTIONS": "true", "DOCKER_ACTIONS_CONTAINERS": "homarr, Dashdot"}))
	if err != nil || len(cfg.DockerActionsContainers) != 2 || cfg.DockerActionsContainers[1] != "Dashdot" {
		t.Fatalf("unexpected %v %v", cfg.DockerActionsContainers, err)
	}
}
