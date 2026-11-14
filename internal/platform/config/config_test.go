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

func TestMetricsNeedAToken(t *testing.T) {
	if _, err := Load(env(map[string]string{"SECRET_KEY": validKey, "METRICS_ENABLED": "true", "METRICS_TOKEN": "short"})); err == nil {
		t.Fatal("short token accepted")
	}
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey, "METRICS_ENABLED": "true", "METRICS_TOKEN": strings.Repeat("a", 32)}))
	if err != nil || !cfg.MetricsEnabled {
		t.Fatalf("%v", err)
	}
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

func TestSmartHomeActionEntities(t *testing.T) {
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey, "SMARTHOME_ACTIONS_ENTITIES": "switch.Coffee, button.door_bell"}))
	if err != nil || len(cfg.SmartHomeActionsEntities) != 2 || cfg.SmartHomeActionsEntities[0] != "switch.coffee" {
		t.Fatalf("unexpected %v %v", cfg.SmartHomeActionsEntities, err)
	}
	if _, err := Load(env(map[string]string{"SECRET_KEY": validKey, "SMARTHOME_ACTIONS_ENTITIES": "coffee"})); err == nil {
		t.Fatal("an ID without domain was accepted")
	}
	if _, err := Load(env(map[string]string{"SECRET_KEY": validKey, "DEMO_MODE": "true", "SMARTHOME_ACTIONS_ENTITIES": "switch.x"})); err == nil {
		t.Fatal("a demo must not switch real devices")
	}
}

func TestProxmoxActionGuests(t *testing.T) {
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey, "PROXMOX_ACTIONS_GUESTS": "100, 105"}))
	if err != nil || len(cfg.ProxmoxActionsGuests) != 2 || cfg.ProxmoxActionsGuests[1] != 105 {
		t.Fatalf("unexpected %v %v", cfg.ProxmoxActionsGuests, err)
	}
	for _, bad := range []string{"vm100", "0", "-1"} {
		if _, err := Load(env(map[string]string{"SECRET_KEY": validKey, "PROXMOX_ACTIONS_GUESTS": bad})); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if _, err := Load(env(map[string]string{"SECRET_KEY": validKey, "DEMO_MODE": "true", "PROXMOX_ACTIONS_GUESTS": "100"})); err == nil {
		t.Fatal("a demo must not have Proxmox actions")
	}
}

func TestDemoModeRefusesRealIntegrations(t *testing.T) {
	vars := map[string]string{"SECRET_KEY": validKey, "DEMO_MODE": "true", "SAMPLE_DATA": "false"}
	cfg, err := Load(env(vars))
	if err != nil || !cfg.DemoMode || !cfg.SampleData {
		t.Fatalf("demo: %+v %v", cfg.DemoMode, err)
	}
	vars["DOCKER_ACTIONS"], vars["DOCKER_ACTIONS_CONTAINERS"] = "true", "x"
	if _, err := Load(env(vars)); err == nil {
		t.Fatal("DOCKER_ACTIONS with DEMO_MODE accepted")
	}
}

func TestSSOValidation(t *testing.T) {
	base := map[string]string{"SECRET_KEY": validKey}
	with := func(kv ...string) map[string]string {
		m := map[string]string{}
		for k, v := range base {
			m[k] = v
		}
		for i := 0; i < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return m
	}
	bad := []map[string]string{
		with("OIDC_ISSUER", "https://id.example"),
		with("OIDC_ISSUER", "https://id.example", "OIDC_CLIENT_ID", "c", "OIDC_REDIRECT_URL", "https://h/cb"),
		with("AUTH_PROXY_HEADER", "Remote-User", "AUTH_PROXY_USERS", "me"),
		with("PASSWORD_LOGIN", "false"),
	}
	for _, vars := range bad {
		if _, err := Load(env(vars)); err == nil {
			t.Errorf("accepted %v", vars)
		}
	}
	ok := with("OIDC_ISSUER", "https://id.example", "OIDC_CLIENT_ID", "c", "OIDC_REDIRECT_URL", "https://h/cb",
		"OIDC_ALLOWED_EMAILS", "me@example.org", "PASSWORD_LOGIN", "false")
	if cfg, err := Load(env(ok)); err != nil || cfg.PasswordLogin {
		t.Fatalf("%v %v", cfg.PasswordLogin, err)
	}
}

func TestDeviceModuleOnlyOnAppliances(t *testing.T) {
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey}))
	if err != nil || cfg.ModuleEnabled("device") {
		t.Fatal("device module outside the appliance", err)
	}
	cfg, err = Load(env(map[string]string{"SECRET_KEY": validKey, "INSTALL_METHOD": "appliance"}))
	if err != nil || !cfg.ModuleEnabled("device") || cfg.DeviceDir != "/var/lib/trisse-home/device" {
		t.Fatal(cfg.DeviceDir, err)
	}
}

func TestDNSPauseProviders(t *testing.T) {
	cfg, err := Load(env(map[string]string{"SECRET_KEY": validKey, "DNS_PAUSE_PROVIDERS": "AdGuard, pihole"}))
	if err != nil || len(cfg.DNSPauseProviders) != 2 || cfg.DNSPauseProviders[0] != "adguard" {
		t.Fatalf("%v %v", cfg.DNSPauseProviders, err)
	}
	if _, err := Load(env(map[string]string{"SECRET_KEY": validKey, "DNS_PAUSE_PROVIDERS": "bind"})); err == nil {
		t.Fatal("unknown provider accepted")
	}
	if _, err := Load(env(map[string]string{"SECRET_KEY": validKey, "DEMO_MODE": "true", "DNS_PAUSE_PROVIDERS": "pihole"})); err == nil {
		t.Fatal("a demo must not pause a real DNS filter")
	}
}
