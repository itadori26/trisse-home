package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"
)

// SecretKeyBytes is the required length of the decoded SECRET_KEY (AES-256).
const SecretKeyBytes = 32

// DefaultAllowedCIDRs covers loopback and all private address ranges, which
// is a safe default for a service that must not be reachable from the internet.
const DefaultAllowedCIDRs = "127.0.0.0/8,::1/128,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,fc00::/7"

// Config is the validated runtime configuration.
type Config struct {
	ListenAddr      string
	DatabasePath    string
	LogLevel        slog.Level
	LogFormat       string
	AllowedCIDRs    []netip.Prefix
	TrustedProxies  []netip.Prefix
	CookieSecure    bool
	SessionTTL      time.Duration
	SecretKey       []byte
	SetupToken      string
	DisabledModules []string
	ShutdownTimeout time.Duration
	DockerProxyURL  string
	HostProcPath    string
	HostSysPath     string
	DiskPaths       []string
}

// ModuleEnabled reports whether the module with the given name is enabled.
func (c Config) ModuleEnabled(name string) bool {
	return !slices.Contains(c.DisabledModules, name)
}

// Load reads the configuration using getenv (usually os.Getenv) and returns
// all problems at once so that misconfiguration is fixed in one go.
func Load(getenv func(string) string) (Config, error) {
	l := loader{getenv: getenv}
	cfg := Config{
		ListenAddr:      l.str("LISTEN_ADDR", ":8090"),
		DatabasePath:    l.str("DATABASE_PATH", "data/trisse-home.db"),
		LogLevel:        l.level("LOG_LEVEL", slog.LevelInfo),
		LogFormat:       l.oneOf("LOG_FORMAT", "json", "json", "text"),
		AllowedCIDRs:    l.prefixes("ALLOWED_CIDRS", DefaultAllowedCIDRs),
		TrustedProxies:  l.prefixes("TRUSTED_PROXIES", ""),
		CookieSecure:    l.boolean("COOKIE_SECURE", false),
		SessionTTL:      l.duration("SESSION_TTL", 30*24*time.Hour),
		SecretKey:       l.secretKey("SECRET_KEY"),
		SetupToken:      l.str("SETUP_TOKEN", ""),
		DisabledModules: l.list("MODULES_DISABLED", ""),
		ShutdownTimeout: l.duration("SHUTDOWN_TIMEOUT", 15*time.Second),
		DockerProxyURL:  l.str("DOCKER_PROXY_URL", "http://docker-socket-proxy:2375"),
		HostProcPath:    l.str("HOST_PROC", "/proc"),
		HostSysPath:     l.str("HOST_SYS", "/sys"),
		DiskPaths:       l.rawList("SYSTEM_DISK_PATHS", "/"),
	}
	if len(cfg.AllowedCIDRs) == 0 && l.getenv("ALLOWED_CIDRS") != "" {
		l.fail("ALLOWED_CIDRS", "must contain at least one CIDR")
	}
	if len(cfg.SetupToken) > 0 && len(cfg.SetupToken) < 16 {
		l.fail("SETUP_TOKEN", "must be at least 16 characters long")
	}
	return cfg, errors.Join(l.errs...)
}

type loader struct {
	getenv func(string) string
	errs   []error
}

func (l *loader) fail(key, msg string) {
	l.errs = append(l.errs, fmt.Errorf("%s: %s", key, msg))
}

func (l *loader) str(key, def string) string {
	if v := strings.TrimSpace(l.getenv(key)); v != "" {
		return v
	}
	return def
}

// list returns a lower-cased, comma-separated list (for identifiers).
func (l *loader) list(key, def string) []string {
	out := l.rawList(key, def)
	for i := range out {
		out[i] = strings.ToLower(out[i])
	}
	return out
}

// rawList returns a comma-separated list preserving case (for paths).
func (l *loader) rawList(key, def string) []string {
	var out []string
	for part := range strings.SplitSeq(l.str(key, def), ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (l *loader) oneOf(key, def string, allowed ...string) string {
	v := strings.ToLower(l.str(key, def))
	if !slices.Contains(allowed, v) {
		l.fail(key, fmt.Sprintf("must be one of %v", allowed))
	}
	return v
}

func (l *loader) boolean(key string, def bool) bool {
	raw := l.str(key, strconv.FormatBool(def))
	v, err := strconv.ParseBool(raw)
	if err != nil {
		l.fail(key, "must be true or false")
	}
	return v
}

func (l *loader) duration(key string, def time.Duration) time.Duration {
	v, err := time.ParseDuration(l.str(key, def.String()))
	if err != nil || v <= 0 {
		l.fail(key, "must be a positive duration such as 720h")
		return def
	}
	return v
}

func (l *loader) level(key string, def slog.Level) slog.Level {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(l.str(key, def.String()))); err != nil {
		l.fail(key, "must be debug, info, warn or error")
		return def
	}
	return lvl
}

func (l *loader) prefixes(key, def string) []netip.Prefix {
	var out []netip.Prefix
	for _, raw := range l.list(key, def) {
		p, err := netip.ParsePrefix(raw)
		if err != nil {
			l.fail(key, fmt.Sprintf("invalid CIDR %q", raw))
			continue
		}
		out = append(out, p.Masked())
	}
	return out
}

func (l *loader) secretKey(key string) []byte {
	raw := l.str(key, "")
	if raw == "" {
		l.fail(key, "is required; generate one with: openssl rand -base64 32")
		return nil
	}
	b, err := base64.StdEncoding.DecodeString(raw)
	if err != nil || len(b) != SecretKeyBytes {
		l.fail(key, fmt.Sprintf("must be %d random bytes, base64 encoded (openssl rand -base64 32)", SecretKeyBytes))
		return nil
	}
	return b
}
