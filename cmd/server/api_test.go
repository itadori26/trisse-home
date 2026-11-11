package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"trisse-home/internal/platform/config"
	"trisse-home/internal/platform/crypto"
)

// flakyDown switches the /flaky endpoint of the fake upstream to 503.
var flakyDown atomic.Bool

// ntfyMessages records messages published to the fake ntfy endpoint.
var ntfyMessages = struct {
	sync.Mutex
	titles []string
}{}

// fakeUpstream emulates every external system the modules talk to.
func fakeUpstream(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /flaky", func(w http.ResponseWriter, _ *http.Request) {
		if flakyDown.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})
	mux.HandleFunc("POST /ntfy", func(w http.ResponseWriter, r *http.Request) {
		var m struct {
			Title string `json:"title"`
		}
		_ = json.NewDecoder(r.Body).Decode(&m)
		ntfyMessages.Lock()
		ntfyMessages.titles = append(ntfyMessages.titles, m.Title)
		ntfyMessages.Unlock()
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /feed", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `<?xml version="1.0"?><rss version="2.0"><channel><title>Fake News</title><link>https://news.example/</link>
			<item><guid>1</guid><title>Hello</title><link>https://news.example/1</link><pubDate>Mon, 09 Nov 2026 10:00:00 GMT</pubDate></item>
			<item><guid>2</guid><title>World</title><link>https://news.example/2</link><pubDate>Tue, 10 Nov 2026 10:00:00 GMT</pubDate></item>
			</channel></rss>`)
	})
	mux.HandleFunc("GET /containers/json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `[{"Id":"abcdef1234567890","Names":["/web"],"Image":"nginx","State":"running","Status":"Up 1 hour (healthy)","Created":1700000000}]`)
	})
	mux.HandleFunc("POST /containers/web/restart", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /containers/web/logs", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "started\nlistening on :80\n")
	})
	mux.HandleFunc("GET /control/stats", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"num_dns_queries":100,"num_blocked_filtering":25,"avg_processing_time":0.002,"top_blocked_domains":[{"ads.example":7}]}`)
	})
	mux.HandleFunc("GET /control/status", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"protection_enabled":true}`)
	})
	mux.HandleFunc("/dav/calendars/me/personal/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "REPORT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		start := time.Now().Add(2 * time.Hour).UTC().Format("20060102T150405Z")
		end := time.Now().Add(3 * time.Hour).UTC().Format("20060102T150405Z")
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = io.WriteString(w, `<d:multistatus xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav"><d:response><d:propstat><d:prop><c:calendar-data>`+
			"BEGIN:VCALENDAR\nBEGIN:VEVENT\nUID:e1\nSUMMARY:Dentist\nDTSTART:"+start+"\nDTEND:"+end+"\nEND:VEVENT\nEND:VCALENDAR\n"+
			`</c:calendar-data></d:prop></d:propstat></d:response></d:multistatus>`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// fakeProc writes a minimal procfs so the system module is deterministic.
func fakeProc(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"stat":    "cpu  100 0 100 700 100 0 0 0 0 0\ncpu0 100 0 100 700 100 0 0 0 0 0\n",
		"loadavg": "0.10 0.20 0.30 1/100 1\n",
		"meminfo": "MemTotal: 1000 kB\nMemAvailable: 500 kB\nSwapTotal: 0 kB\nSwapFree: 0 kB\n",
		"uptime":  "120.0 100.0\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

type client struct {
	t    *testing.T
	base string
	http *http.Client
	csrf string
}

func (c *client) do(method, path string, body any, wantStatus int) []byte {
	c.t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, _ := http.NewRequestWithContext(context.Background(), method, c.base+path, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	}
	res, err := c.http.Do(req)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != wantStatus {
		c.t.Fatalf("%s %s: status %d, want %d: %s", method, path, res.StatusCode, wantStatus, out)
	}
	return out
}

func (c *client) json(method, path string, body any, wantStatus int, dst any) {
	c.t.Helper()
	out := c.do(method, path, body, wantStatus)
	if dst != nil {
		if err := json.Unmarshal(out, dst); err != nil {
			c.t.Fatalf("%s %s: decode %q: %v", method, path, out, err)
		}
	}
}

type obj = map[string]any

func newTestServer(t *testing.T, upstream string, opts ...func(*config.Config)) *httptest.Server {
	t.Helper()
	key := make([]byte, config.SecretKeyBytes)
	_, _ = rand.Read(key)
	cfg := config.Config{
		ListenAddr: ":0", DatabasePath: filepath.Join(t.TempDir(), "e2e.db"),
		AllowedCIDRs: []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}, SessionTTL: time.Hour,
		SecretKey: key, SetupToken: "e2e-setup-token-0123", DockerProxyURL: upstream,
		DockerActions: true, DockerActionsContainers: []string{"web"}, DockerActionsProxyURL: upstream,
		HostProcPath: fakeProc(t), HostSysPath: t.TempDir(), DiskPaths: []string{"Temp=" + t.TempDir()},
		ShutdownTimeout: time.Second,
		BackupDir:       filepath.Join(t.TempDir(), "backups"), BackupInterval: 24 * time.Hour, BackupRetention: 3,
	}
	for _, o := range opts {
		o(&cfg)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	a, err := build(context.Background(), cfg, log)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.db.Close() })
	srv := httptest.NewServer(newHandler(a))
	t.Cleanup(srv.Close)
	return srv
}

func newClient(t *testing.T, base string) *client {
	jar, _ := cookiejar.New(nil)
	return &client{t: t, base: base, http: &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}}
}

func TestAPIEndToEnd(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	c := newClient(t, srv.URL)

	// --- platform and auth ---
	c.do(http.MethodGet, "/healthz", nil, http.StatusOK)
	c.do(http.MethodGet, "/api/nope", nil, http.StatusNotFound)
	var status obj
	c.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &status)
	if status["setupRequired"] != true {
		t.Fatalf("expected setup required: %v", status)
	}
	c.do(http.MethodGet, "/api/dashboard", nil, http.StatusUnauthorized)
	c.do(http.MethodPost, "/api/auth/setup", obj{"setupToken": "wrong", "password": "e2e-password-123"}, http.StatusUnauthorized)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.do(http.MethodPost, "/api/dashboard/tabs", obj{"name": "x"}, http.StatusForbidden) // no CSRF token yet
	c.csrf = session["csrfToken"].(string)

	var info obj
	c.json(http.MethodGet, "/api/modules", nil, http.StatusOK, &info)
	if n := len(info["modules"].([]any)); n != 18 {
		t.Fatalf("expected 18 modules, got %d: %v", n, info)
	}

	// --- backups ---
	var backupFile obj
	c.json(http.MethodPost, "/api/backups", nil, http.StatusCreated, &backupFile)
	var backups []obj
	c.json(http.MethodGet, "/api/backups", nil, http.StatusOK, &backups)
	if len(backups) != 1 {
		t.Fatalf("expected one backup, got %v", backups)
	}
	if body := c.do(http.MethodGet, "/api/backups/"+backupFile["name"].(string), nil, http.StatusOK); !strings.HasPrefix(string(body), "SQLite format 3") {
		t.Fatal("backup download is not an SQLite database")
	}
	c.do(http.MethodGet, "/api/backups/..%2F..%2Fetc%2Fpasswd", nil, http.StatusNotFound)

	// --- onboarding ---
	var tour obj
	c.json(http.MethodGet, "/api/onboarding", nil, http.StatusOK, &tour)
	if tour["completed"] != false {
		t.Fatalf("tour must be pending after setup: %v", tour)
	}
	c.json(http.MethodPut, "/api/onboarding", obj{"completed": true}, http.StatusOK, &tour)
	if tour["completed"] != true {
		t.Fatalf("tour not completed: %v", tour)
	}

	// --- dashboard ---
	var dash struct {
		Tabs []struct {
			ID      string `json:"id"`
			Widgets []obj  `json:"widgets"`
		} `json:"tabs"`
	}
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	if len(dash.Tabs) != 1 || len(dash.Tabs[0].Widgets) != 2 {
		t.Fatalf("default dashboard wrong: %+v", dash)
	}
	home := dash.Tabs[0].ID
	var tab, widget obj
	c.json(http.MethodPost, "/api/dashboard/tabs", obj{"name": "Server"}, http.StatusCreated, &tab)
	tabID := tab["id"].(string)
	c.json(http.MethodPatch, "/api/dashboard/tabs/"+tabID, obj{"name": "Servers"}, http.StatusOK, nil)
	c.do(http.MethodPut, "/api/dashboard/tabs/order", obj{"tabIds": []string{tabID, home}}, http.StatusNoContent)
	c.do(http.MethodPost, "/api/dashboard/tabs/"+tabID+"/widgets", obj{"kind": "unknown", "w": 2, "h": 2}, http.StatusUnprocessableEntity)
	c.json(http.MethodPost, "/api/dashboard/tabs/"+tabID+"/widgets", obj{"kind": "system", "x": 0, "y": 0, "w": 4, "h": 4}, http.StatusCreated, &widget)
	wID := widget["id"].(string)
	c.do(http.MethodPut, "/api/dashboard/tabs/"+tabID+"/layout", obj{"items": []obj{{"id": wID, "x": 2, "y": 0, "w": 5, "h": 4}}}, http.StatusNoContent)
	c.json(http.MethodPatch, "/api/dashboard/widgets/"+wID, obj{"config": obj{"a": 1}}, http.StatusOK, nil)
	c.do(http.MethodDelete, "/api/dashboard/widgets/"+wID, nil, http.StatusNoContent)
	c.do(http.MethodDelete, "/api/dashboard/tabs/"+tabID, nil, http.StatusNoContent)
	c.do(http.MethodDelete, "/api/dashboard/tabs/"+home, nil, http.StatusConflict) // last tab

	// --- appearance ---
	c.do(http.MethodPut, "/api/appearance", obj{"theme": "neon", "accentColor": "", "backgroundImage": "", "customCss": ""}, http.StatusUnprocessableEntity)
	c.json(http.MethodPut, "/api/appearance", obj{"theme": "dark", "accentColor": "#112233", "backgroundImage": "", "customCss": ""}, http.StatusOK, nil)
	c.json(http.MethodGet, "/api/appearance", nil, http.StatusOK, nil)

	// --- search ---
	c.do(http.MethodGet, "/api/search?q=hello", nil, http.StatusFound)
	c.json(http.MethodPut, "/api/search/settings", obj{"defaultProvider": "brave"}, http.StatusOK, nil)
	c.json(http.MethodPost, "/api/search/shortcuts", obj{"keyword": "gh", "name": "GitHub", "urlTemplate": "https://github.com/search?q={query}"}, http.StatusOK, nil)
	c.json(http.MethodGet, "/api/search/settings", nil, http.StatusOK, nil)
	c.json(http.MethodDelete, "/api/search/shortcuts/gh", nil, http.StatusOK, nil)
	if body := string(c.do(http.MethodGet, "/api/search/opensearch.xml", nil, http.StatusOK)); !strings.Contains(body, "OpenSearchDescription") {
		t.Fatal("opensearch document missing")
	}

	// --- weather (validation only; no outbound calls in tests) ---
	c.do(http.MethodGet, "/api/weather?lat=999&lon=1", nil, http.StatusUnprocessableEntity)
	c.do(http.MethodGet, "/api/weather/places?q=x", nil, http.StatusUnprocessableEntity)

	// --- services ---
	var svc obj
	c.json(http.MethodPost, "/api/services", obj{"name": "Fake", "url": up.URL + "/health"}, http.StatusCreated, &svc)
	if svc["state"] != "online" {
		t.Fatalf("service not online: %v", svc)
	}
	sID := svc["id"].(string)
	c.json(http.MethodPut, "/api/services/"+sID, obj{"name": "Fake2", "url": up.URL + "/health"}, http.StatusOK, nil)
	c.do(http.MethodPut, "/api/services/order", obj{"ids": []string{sID}}, http.StatusNoContent)
	c.json(http.MethodPost, "/api/services/check", nil, http.StatusOK, nil)
	c.json(http.MethodGet, "/api/services", nil, http.StatusOK, nil)
	c.do(http.MethodDelete, "/api/services/"+sID, nil, http.StatusNoContent)

	// --- notifications: a service going offline is pushed via ntfy ---
	rules := obj{"services": true, "containers": true, "disk": true, "diskThreshold": 90, "tls": true, "tlsDays": 14, "backups": true}
	c.json(http.MethodPut, "/api/notifications/settings", obj{"enabled": true, "serverUrl": up.URL + "/ntfy", "topic": "home", "token": "", "rules": rules}, http.StatusOK, nil)
	c.do(http.MethodPost, "/api/notifications/test", nil, http.StatusNoContent)
	var flaky obj
	c.json(http.MethodPost, "/api/services", obj{"name": "Flaky", "url": up.URL + "/flaky"}, http.StatusCreated, &flaky)
	flakyDown.Store(true)
	c.json(http.MethodPost, "/api/services/check", nil, http.StatusOK, nil)
	c.json(http.MethodPost, "/api/services/check", nil, http.StatusOK, nil)
	waitFor(t, func() bool {
		ntfyMessages.Lock()
		defer ntfyMessages.Unlock()
		return slices.Contains(ntfyMessages.titles, "Flaky is offline")
	})
	c.do(http.MethodDelete, "/api/services/"+flaky["id"].(string), nil, http.StatusNoContent)

	// --- backup monitor: external scripts report via a public ping URL ---
	var created struct {
		Source  obj    `json:"source"`
		PingURL string `json:"pingUrl"`
	}
	c.json(http.MethodPost, "/api/backup-status/sources", obj{"name": "NAS", "maxAgeHours": 26}, http.StatusCreated, &created)
	anonymous := newClient(t, srv.URL)
	pingPath := strings.TrimPrefix(created.PingURL, srv.URL)
	anonymous.do(http.MethodPost, pingPath, nil, http.StatusNoContent)
	anonymous.do(http.MethodPost, "/api/backup-status/ping/wrong-token", nil, http.StatusNotFound)
	var sources []obj
	c.json(http.MethodGet, "/api/backup-status/sources", nil, http.StatusOK, &sources)
	if len(sources) != 1 || sources[0]["state"] != "ok" {
		t.Fatalf("backup source: %v", sources)
	}

	// --- bookmarks ---
	var folder, bm obj
	c.json(http.MethodPost, "/api/bookmarks/folders", obj{"name": "Dev"}, http.StatusCreated, &folder)
	fID := folder["id"].(string)
	c.json(http.MethodPost, "/api/bookmarks", obj{"url": "https://go.dev", "folderId": fID, "tags": []string{"go"}}, http.StatusCreated, &bm)
	var list struct {
		Bookmarks []obj `json:"bookmarks"`
	}
	c.json(http.MethodGet, "/api/bookmarks?tag=go", nil, http.StatusOK, &list)
	if len(list.Bookmarks) != 1 {
		t.Fatalf("bookmark filter wrong: %+v", list)
	}
	c.json(http.MethodPut, "/api/bookmarks/"+bm["id"].(string), obj{"url": "https://go.dev/doc", "title": "Docs"}, http.StatusOK, nil)
	c.json(http.MethodPatch, "/api/bookmarks/folders/"+fID, obj{"name": "Development"}, http.StatusOK, nil)
	c.do(http.MethodDelete, "/api/bookmarks/folders/"+fID, nil, http.StatusNoContent)
	c.do(http.MethodDelete, "/api/bookmarks/"+bm["id"].(string), nil, http.StatusNoContent)

	// --- notes ---
	var note obj
	c.json(http.MethodPost, "/api/notes", obj{"title": "", "body": "# Hello", "pinned": true}, http.StatusCreated, &note)
	nID := note["id"].(string)
	c.json(http.MethodPut, "/api/notes/"+nID, obj{"title": "Hi", "body": "x", "pinned": false}, http.StatusOK, nil)
	c.json(http.MethodGet, "/api/notes/"+nID, nil, http.StatusOK, nil)
	c.json(http.MethodGet, "/api/notes?pinned=true", nil, http.StatusOK, nil)
	c.do(http.MethodDelete, "/api/notes/"+nID, nil, http.StatusNoContent)

	// --- to-dos ---
	var todo obj
	c.json(http.MethodPost, "/api/todos", obj{"title": "Test", "due": "2026-12-24"}, http.StatusCreated, &todo)
	c.json(http.MethodPut, "/api/todos/"+todo["id"].(string), obj{"title": "Test", "due": "", "done": true}, http.StatusOK, nil)
	c.json(http.MethodGet, "/api/todos?done=true", nil, http.StatusOK, nil)
	var cleared obj
	c.json(http.MethodPost, "/api/todos/clear-completed", nil, http.StatusOK, &cleared)
	if cleared["removed"] != float64(1) {
		t.Fatalf("clear completed: %v", cleared)
	}

	// --- rss ---
	var feed obj
	c.json(http.MethodPost, "/api/rss/feeds", obj{"url": up.URL + "/feed"}, http.StatusCreated, &feed)
	if feed["title"] != "Fake News" || feed["unread"] != float64(2) {
		t.Fatalf("feed: %v", feed)
	}
	var items []obj
	c.json(http.MethodGet, "/api/rss/items?unread=true", nil, http.StatusOK, &items)
	c.do(http.MethodPut, "/api/rss/items/"+items[0]["id"].(string)+"/read", obj{"read": true}, http.StatusNoContent)
	c.do(http.MethodPost, "/api/rss/read-all", obj{"feedId": ""}, http.StatusNoContent)
	c.json(http.MethodPost, "/api/rss/refresh", nil, http.StatusOK, nil)
	c.do(http.MethodDelete, "/api/rss/feeds/"+feed["id"].(string), nil, http.StatusNoContent)

	// --- system, docker ---
	var sys obj
	c.json(http.MethodGet, "/api/system", nil, http.StatusOK, &sys)
	if sys["cpuCount"] != float64(1) || sys["memory"].(obj)["percent"] != float64(50) {
		t.Fatalf("system: %v", sys)
	}
	var containers obj
	c.json(http.MethodGet, "/api/docker/containers", nil, http.StatusOK, &containers)
	if containers["running"] != float64(1) {
		t.Fatalf("docker: %v", containers)
	}
	c.do(http.MethodPost, "/api/docker/containers/web/restart", nil, http.StatusNoContent)
	c.do(http.MethodPost, "/api/docker/containers/other/restart", nil, http.StatusForbidden)
	var logs struct{ Lines []string }
	c.json(http.MethodGet, "/api/docker/containers/web/logs", nil, http.StatusOK, &logs)
	if len(logs.Lines) != 2 || logs.Lines[1] != "listening on :80" {
		t.Fatalf("docker logs: %v", logs)
	}

	// --- adguard ---
	c.do(http.MethodGet, "/api/adguard/stats", nil, http.StatusNotFound)
	c.json(http.MethodPut, "/api/adguard/settings", obj{"url": up.URL, "username": "admin", "password": "pw"}, http.StatusOK, nil)
	if body := string(c.do(http.MethodGet, "/api/adguard/settings", nil, http.StatusOK)); strings.Contains(body, `"pw"`) {
		t.Fatal("password leaked")
	}
	var stats obj
	c.json(http.MethodGet, "/api/adguard/stats", nil, http.StatusOK, &stats)
	if stats["blockedPercent"] != float64(25) {
		t.Fatalf("adguard: %v", stats)
	}

	// --- calendar ---
	c.do(http.MethodGet, "/api/calendar/events", nil, http.StatusNotFound)
	c.json(http.MethodPut, "/api/calendar/settings", obj{"calendarUrls": []string{up.URL + "/dav/calendars/me/personal/"}, "username": "me", "password": "pw"}, http.StatusOK, nil)
	c.json(http.MethodGet, "/api/calendar/settings", nil, http.StatusOK, nil)
	var agenda struct {
		Events []obj `json:"events"`
	}
	c.json(http.MethodGet, "/api/calendar/events?days=7", nil, http.StatusOK, &agenda)
	if len(agenda.Events) != 1 || agenda.Events[0]["summary"] != "Dentist" {
		t.Fatalf("calendar: %+v", agenda)
	}

	// --- sessions ---
	other := newClient(t, srv.URL)
	other.json(http.MethodPost, "/api/auth/login", obj{"password": "e2e-password-123"}, http.StatusOK, nil)
	var sessions []obj
	c.json(http.MethodGet, "/api/auth/sessions", nil, http.StatusOK, &sessions)
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %v", sessions)
	}
	var revoked obj
	c.json(http.MethodPost, "/api/auth/sessions/revoke-others", nil, http.StatusOK, &revoked)
	if revoked["revoked"] != float64(1) {
		t.Fatalf("revoke others: %v", revoked)
	}
	other.do(http.MethodGet, "/api/dashboard", nil, http.StatusUnauthorized)
	c.do(http.MethodDelete, "/api/auth/sessions/unknown", nil, http.StatusNotFound)

	// --- two-factor authentication ---
	var enrollment obj
	c.json(http.MethodPost, "/api/auth/totp/setup", nil, http.StatusOK, &enrollment)
	code, err := crypto.TOTP{}.Code(enrollment["secret"].(string), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	var recovery struct {
		Codes []string `json:"codes"`
	}
	c.json(http.MethodPost, "/api/auth/totp/confirm", obj{"code": code}, http.StatusOK, &recovery)
	if len(recovery.Codes) != 10 {
		t.Fatalf("expected 10 recovery codes, got %v", recovery.Codes)
	}
	second := newClient(t, srv.URL)
	if body := string(second.do(http.MethodPost, "/api/auth/login", obj{"password": "e2e-password-123"}, http.StatusUnprocessableEntity)); !strings.Contains(body, `"field":"totpCode"`) {
		t.Fatalf("login without second factor: %s", body)
	}
	second.do(http.MethodPost, "/api/auth/login", obj{"password": "e2e-password-123", "code": "000000"}, http.StatusUnauthorized)
	second.json(http.MethodPost, "/api/auth/login", obj{"password": "e2e-password-123", "code": recovery.Codes[0]}, http.StatusOK, nil)
	var tfStatus obj
	c.json(http.MethodGet, "/api/auth/totp", nil, http.StatusOK, &tfStatus)
	if tfStatus["enabled"] != true || tfStatus["recoveryCodesLeft"] != float64(9) {
		t.Fatalf("totp status: %v", tfStatus)
	}
	c.do(http.MethodPost, "/api/auth/totp/disable", obj{"password": "e2e-password-123", "code": recovery.Codes[1]}, http.StatusNoContent)

	// --- password change, logout, login ---
	var changed obj
	c.json(http.MethodPut, "/api/auth/password", obj{"currentPassword": "e2e-password-123", "newPassword": "e2e-password-456"}, http.StatusOK, &changed)
	c.csrf = changed["csrfToken"].(string)
	c.do(http.MethodPost, "/api/auth/logout", nil, http.StatusNoContent)
	c.do(http.MethodGet, "/api/dashboard", nil, http.StatusUnauthorized)
	c.csrf = ""
	c.do(http.MethodPost, "/api/auth/login", obj{"password": "e2e-password-123"}, http.StatusUnauthorized)
	c.json(http.MethodPost, "/api/auth/login", obj{"password": "e2e-password-456"}, http.StatusOK, nil)
	c.do(http.MethodGet, "/api/dashboard", nil, http.StatusOK)
}

func TestCrossSiteRequestsAreRejected(t *testing.T) {
	srv := newTestServer(t, "http://127.0.0.1:1")
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/auth/login", strings.NewReader(`{"password":"x"}`))
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-site login not rejected: %d", res.StatusCode)
	}
}

func TestSampleDataAfterSetup(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) {
		c.SampleData = true
		c.SampleFeeds = []string{up.URL + "/feed"}
		c.SampleWeather = config.SampleLocation{Name: "Oldenburg", Lat: 53.14, Lon: 8.21}
	})
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	var dash struct {
		Tabs []struct {
			Name    string `json:"name"`
			Widgets []struct {
				Kind   string `json:"kind"`
				Config obj    `json:"config"`
			} `json:"widgets"`
		} `json:"tabs"`
	}
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	if len(dash.Tabs) != 2 || dash.Tabs[0].Name != "Home" || dash.Tabs[1].Name != "Server" {
		t.Fatalf("unexpected tabs %+v", dash.Tabs)
	}
	kinds := map[string]obj{}
	for _, tab := range dash.Tabs {
		for _, w := range tab.Widgets {
			kinds[w.Kind] = w.Config
		}
	}
	for _, k := range []string{"search", "clock", "weather", "todos", "notes", "bookmarks", "rss", "calendar", "services", "system", "docker", "adguard"} {
		if _, ok := kinds[k]; !ok {
			t.Errorf("sample dashboard lacks widget %q", k)
		}
	}
	if kinds["weather"]["label"] != "Oldenburg" {
		t.Errorf("weather widget not configured: %v", kinds["weather"])
	}

	var notes []obj
	c.json(http.MethodGet, "/api/notes?pinned=true", nil, http.StatusOK, &notes)
	var todos []obj
	c.json(http.MethodGet, "/api/todos", nil, http.StatusOK, &todos)
	var bookmarks struct {
		Bookmarks []obj `json:"bookmarks"`
		Folders   []obj `json:"folders"`
	}
	c.json(http.MethodGet, "/api/bookmarks", nil, http.StatusOK, &bookmarks)
	var services []obj
	c.json(http.MethodGet, "/api/services", nil, http.StatusOK, &services)
	if len(notes) != 1 || len(todos) != 4 || len(bookmarks.Bookmarks) != 6 || len(bookmarks.Folders) != 2 || len(services) != 3 {
		t.Fatalf("sample data incomplete: notes=%d todos=%d bookmarks=%d folders=%d services=%d",
			len(notes), len(todos), len(bookmarks.Bookmarks), len(bookmarks.Folders), len(services))
	}
	c.do(http.MethodGet, "/api/search?q=!maps+Berlin", nil, http.StatusFound)

	// The example feed is fetched in the background right after setup.
	deadline := time.Now().Add(5 * time.Second)
	for {
		var items []obj
		c.json(http.MethodGet, "/api/rss/items", nil, http.StatusOK, &items)
		if len(items) == 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("sample feed not fetched, %d items", len(items))
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestSampleDataCanBeDisabled(t *testing.T) {
	srv := newTestServer(t, "http://127.0.0.1:1") // SampleData is false in the base config
	c := newClient(t, srv.URL)
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, nil)
	var notes []obj
	c.json(http.MethodGet, "/api/notes", nil, http.StatusOK, &notes)
	if len(notes) != 0 {
		t.Fatalf("expected no sample data, got %d notes", len(notes))
	}
}

// waitFor polls cond for up to five seconds.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if cond() {
			return
		}
	}
	t.Fatal("condition not met within 5 seconds")
}
