package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
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

// wolTarget receives the magic packets of the end-to-end test.
var wolTarget net.PacketConn

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
		_, _ = io.WriteString(w, `[{"Id":"abcdef1234567890","Names":["/web"],"Image":"nginx","State":"running","Status":"Up 1 hour (healthy)","Created":1700000000,`+
			`"Labels":{"homepage.name":"Web","homepage.href":"https://web.lan"}}]`)
	})
	mux.HandleFunc("POST /containers/web/restart", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /containers/web/logs", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "started\nlistening on :80\n")
	})
	tasks := map[string]string{}
	var tasksMu sync.Mutex
	mux.HandleFunc("/dav/tasks/", func(w http.ResponseWriter, r *http.Request) {
		tasksMu.Lock()
		defer tasksMu.Unlock()
		name := strings.TrimPrefix(r.URL.Path, "/dav/tasks/")
		switch r.Method {
		case "REPORT":
			w.WriteHeader(http.StatusMultiStatus)
			var b strings.Builder
			b.WriteString(`<d:multistatus xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">`)
			for n, data := range tasks {
				b.WriteString(`<d:response><d:href>/dav/tasks/` + n + `</d:href><d:propstat><d:prop><c:calendar-data>` + data + `</c:calendar-data></d:prop></d:propstat></d:response>`)
			}
			b.WriteString(`</d:multistatus>`)
			_, _ = io.WriteString(w, b.String())
		case http.MethodGet:
			_, _ = io.WriteString(w, tasks[name])
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			tasks[name] = string(body)
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			delete(tasks, name)
			w.WriteHeader(http.StatusNoContent)
		}
	})
	mux.HandleFunc("GET /api/v1/geocode", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `[{"type":"STOP","id":"de:03403:16750","name":"Oldenburg Hbf","lat":53.14,"lon":8.22}]`)
	})
	mux.HandleFunc("GET /api/v1/stoptimes", func(w http.ResponseWriter, _ *http.Request) {
		planned := time.Now().Add(5 * time.Minute).UTC().Format(time.RFC3339)
		expected := time.Now().Add(8 * time.Minute).UTC().Format(time.RFC3339)
		_, _ = io.WriteString(w, `{"stopTimes":[{"place":{"departure":"`+expected+`","scheduledDeparture":"`+planned+`","track":"A"},"mode":"BUS","realTime":true,"headsign":"Bümmerstede","routeShortName":"304"}]}`)
	})
	mux.HandleFunc("GET /abfall.ics", func(w http.ResponseWriter, _ *http.Request) {
		tomorrow := time.Now().AddDate(0, 0, 1).Format("20060102")
		_, _ = io.WriteString(w, "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:w1\r\nSUMMARY:Biotonne\r\nDTSTART;VALUE=DATE:"+tomorrow+
			"\r\nRRULE:FREQ=WEEKLY;INTERVAL=2\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n")
	})
	mux.HandleFunc("GET /status.php", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"installed":true,"maintenance":false,"versionstring":"30.0.2"}`)
	})
	mux.HandleFunc("GET /ocs/v1.php/cloud/users/anna", func(w http.ResponseWriter, r *http.Request) {
		if _, p, _ := r.BasicAuth(); p != "app-pw" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = io.WriteString(w, `{"ocs":{"meta":{"statuscode":100},"data":{"quota":{"used":250,"free":750,"quota":1000}}}}`)
	})
	mux.HandleFunc("GET /control/stats", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"num_dns_queries":100,"num_blocked_filtering":25,"avg_processing_time":0.002,"top_blocked_domains":[{"ads.example":7}]}`)
	})
	mux.HandleFunc("GET /control/status", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"protection_enabled":true}`)
	})
	// Sonarr (#86).
	mux.HandleFunc("GET /sonarr/api/v3/system/status", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"appName":"Sonarr"}`)
	})
	mux.HandleFunc("GET /sonarr/api/v3/calendar", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `[{"title":"Pilot","seasonNumber":1,"episodeNumber":1,"airDateUtc":"`+time.Now().Add(24*time.Hour).UTC().Format(time.RFC3339)+`","series":{"title":"Show"}}]`)
	})
	mux.HandleFunc("GET /sonarr/api/v3/queue", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, `{"records":[]}`) })
	mux.HandleFunc("GET /sonarr/api/v3/health", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, `[]`) })
	mux.HandleFunc("GET /sonarr/api/v3/wanted/missing", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"totalRecords":2}`)
	})
	mux.HandleFunc("GET /ha/api/states", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer ha-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = io.WriteString(w, `[{"entity_id":"switch.coffee","state":"off","attributes":{"friendly_name":"Coffee"},"last_changed":"2026-11-12T10:00:00Z"}]`)
	})
	mux.HandleFunc("GET /api2/json/cluster/resources", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "PVEAPIToken=monitor@pve!th=uuid" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = io.WriteString(w, `{"data":[{"type":"node","node":"pve","status":"online","cpu":0.5,"maxcpu":4},`+
			`{"type":"qemu","vmid":100,"name":"ha","node":"pve","status":"running"}]}`)
	})
	// Pi-hole v6 (#83) without a password: no session needed.
	mux.HandleFunc("GET /pihole/api/stats/summary", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"queries":{"total":200,"blocked":20}}`)
	})
	mux.HandleFunc("GET /pihole/api/dns/blocking", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"blocking":"enabled"}`)
	})
	mux.HandleFunc("GET /pihole/api/stats/top_domains", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"domains":[{"domain":"ads.example","count":9}]}`)
	})
	mux.HandleFunc("/dav/calendars/me/personal/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.Header.Get("If-None-Match") == "*" && strings.HasSuffix(r.URL.Path, ".ics") {
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "SUMMARY:Haircut") {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		}
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
	var err error
	wolTarget, err = net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = wolTarget.Close() })
	key := make([]byte, config.SecretKeyBytes)
	_, _ = rand.Read(key)
	cfg := config.Config{
		ListenAddr: ":0", DatabasePath: filepath.Join(t.TempDir(), "e2e.db"),
		AllowedCIDRs: []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}, SessionTTL: time.Hour,
		SecretKey: key, SetupToken: "e2e-setup-token-0123", DockerProxyURL: upstream,
		DockerActions: true, DockerActionsContainers: []string{"web"}, DockerActionsProxyURL: upstream,
		WolBroadcast: wolTarget.LocalAddr().String(), TransitAPIURL: upstream,
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
	// Every module is enabled by default; spot-check core and optional ones
	// instead of a count, so adding a module does not break this test.
	enabled := map[string]bool{}
	for _, m := range info["modules"].([]any) {
		enabled[m.(string)] = true
	}
	for _, name := range []string{"auth", "dashboard", "system", "backup", "transfer", "calendar", "wireguard"} {
		if !enabled[name] {
			t.Fatalf("module %s missing: %v", name, info)
		}
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
	// templates and per-tab background
	var tpl obj
	c.json(http.MethodPost, "/api/dashboard/templates", obj{"tabId": tabID, "name": "Server set"}, http.StatusCreated, &tpl)
	var fromTpl obj
	c.json(http.MethodPost, "/api/dashboard/templates/"+tpl["id"].(string)+"/tabs", obj{"name": ""}, http.StatusCreated, &fromTpl)
	if fromTpl["name"] != "Server set" || len(fromTpl["widgets"].([]any)) != 1 {
		t.Fatalf("tab from template: %v", fromTpl)
	}
	c.do(http.MethodPut, "/api/dashboard/tabs/"+fromTpl["id"].(string)+"/background", obj{"background": "javascript:x"}, http.StatusUnprocessableEntity)
	c.json(http.MethodPut, "/api/dashboard/tabs/"+fromTpl["id"].(string)+"/background", obj{"background": "https://img.example/a.jpg"}, http.StatusOK, nil)
	c.do(http.MethodDelete, "/api/dashboard/templates/"+tpl["id"].(string), nil, http.StatusNoContent)
	c.do(http.MethodDelete, "/api/dashboard/tabs/"+fromTpl["id"].(string), nil, http.StatusNoContent)
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

	// --- wake-on-lan: magic packet sent directly to WOL_BROADCAST ---
	var nas obj
	c.json(http.MethodPost, "/api/services", obj{"name": "NAS", "url": up.URL + "/flaky", "mac": "AA-BB-CC-DD-EE-FF"}, http.StatusCreated, &nas)
	if nas["mac"] != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("mac not normalised: %v", nas["mac"])
	}
	c.do(http.MethodPost, "/api/services/"+nas["id"].(string)+"/wake", nil, http.StatusNoContent)
	buf := make([]byte, 200)
	_ = wolTarget.SetReadDeadline(time.Now().Add(2 * time.Second))
	if n, _, err := wolTarget.ReadFrom(buf); err != nil || n != 102 {
		t.Fatalf("magic packet not received: %d %v", n, err)
	}
	c.do(http.MethodDelete, "/api/services/"+nas["id"].(string), nil, http.StatusNoContent)

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

	// --- Docker autodiscovery (#78): off by default, sync on demand ---
	var discovery obj
	c.json(http.MethodGet, "/api/services/discovery", nil, http.StatusOK, &discovery)
	if discovery["available"] != true || discovery["enabled"] != false || discovery["lastSync"] != nil {
		t.Fatalf("discovery: %v", discovery)
	}
	c.do(http.MethodPut, "/api/services/discovery", obj{"enabled": true, "intervalMinutes": 0, "rule": "labels"}, http.StatusUnprocessableEntity)
	c.json(http.MethodPost, "/api/services/discovery/sync", nil, http.StatusOK, &discovery)
	if sync := discovery["lastSync"].(obj); sync["created"] != float64(1) || sync["error"] != "" {
		t.Fatalf("discovery sync: %v", discovery)
	}
	var tiles []obj
	c.json(http.MethodGet, "/api/services", nil, http.StatusOK, &tiles)
	web := tiles[len(tiles)-1]
	if web["container"] != "web" || web["name"] != "Web" || web["url"] != "https://web.lan" || web["hidden"] != false {
		t.Fatalf("discovered tile: %v", web)
	}
	c.do(http.MethodDelete, "/api/services/"+web["id"].(string), nil, http.StatusNoContent)

	// --- weather alerts outside Germany need no DWD request ---
	var noWarnings []obj
	c.json(http.MethodGet, "/api/weather/warnings?lat=48.86&lon=2.35", nil, http.StatusOK, &noWarnings)
	if len(noWarnings) != 0 {
		t.Fatalf("warnings outside Germany: %v", noWarnings)
	}
	c.do(http.MethodGet, "/api/weather/radar?lat=48.86&lon=2.35", nil, http.StatusUnprocessableEntity)

	// --- parcels: stored with a hint until a DHL key is configured ---
	var parcel obj
	c.json(http.MethodPost, "/api/parcels", obj{"trackingNumber": "0034 0434 1234 5678", "carrier": "dhl", "label": "Shoes"}, http.StatusCreated, &parcel)
	if parcel["trackingNumber"] != "0034043412345678" || parcel["error"] == "" {
		t.Fatalf("parcel: %v", parcel)
	}
	c.do(http.MethodPost, "/api/parcels", obj{"trackingNumber": "0034043412345678", "carrier": "hermes"}, http.StatusUnprocessableEntity)
	c.do(http.MethodDelete, "/api/parcels/"+parcel["id"].(string), nil, http.StatusNoContent)

	// --- transit departures ---
	var stops []obj
	c.json(http.MethodGet, "/api/transit/stops?q=Oldenburg", nil, http.StatusOK, &stops)
	var deps []obj
	c.json(http.MethodGet, "/api/transit/departures?stop="+stops[0]["id"].(string)+"&n=5", nil, http.StatusOK, &deps)
	if len(deps) != 1 || deps[0]["delayMinutes"] != float64(3) {
		t.Fatalf("transit: %v", deps)
	}

	// --- waste: ICS subscription with a recurring rule ---
	c.do(http.MethodGet, "/api/waste/upcoming", nil, http.StatusNotFound)
	c.json(http.MethodPut, "/api/waste/settings", obj{"url": up.URL + "/abfall.ics", "rules": []obj{}}, http.StatusOK, nil)
	var waste struct{ Types []obj }
	c.json(http.MethodGet, "/api/waste/upcoming", nil, http.StatusOK, &waste)
	if len(waste.Types) != 1 || waste.Types[0]["daysUntil"] != float64(1) || waste.Types[0]["icon"] != "🍂" || len(waste.Types[0]["dates"].([]any)) != 3 {
		t.Fatalf("waste: %+v", waste)
	}

	// --- shopping list on a CalDAV task list ---
	c.json(http.MethodPut, "/api/shopping/settings", obj{"listUrl": up.URL + "/dav/tasks/", "username": "me", "password": "pw"}, http.StatusOK, nil)
	var milk obj
	c.json(http.MethodPost, "/api/shopping/items", obj{"summary": "Milk"}, http.StatusCreated, &milk)
	c.do(http.MethodPut, "/api/shopping/items/"+milk["id"].(string), obj{"done": true}, http.StatusNoContent)
	var shoppingItems []obj
	c.json(http.MethodGet, "/api/shopping/items", nil, http.StatusOK, &shoppingItems)
	if len(shoppingItems) != 1 || shoppingItems[0]["done"] != true {
		t.Fatalf("shopping: %v", shoppingItems)
	}
	c.do(http.MethodPut, "/api/shopping/items/..%2Fsecret.ics", obj{"done": true}, http.StatusNotFound)
	c.json(http.MethodPost, "/api/shopping/clear-completed", nil, http.StatusOK, nil)

	// --- fritzbox: not configured until saved ---
	c.do(http.MethodGet, "/api/fritzbox/status", nil, http.StatusNotFound)
	var fb obj
	c.json(http.MethodGet, "/api/fritzbox/settings", nil, http.StatusOK, &fb)
	if fb["configured"] != false || fb["url"] != "http://192.168.178.1:49000" {
		t.Fatalf("fritzbox defaults: %v", fb)
	}

	// --- nextcloud (no serverinfo app on the fake: degrades gracefully) ---
	c.do(http.MethodGet, "/api/nextcloud/status", nil, http.StatusNotFound)
	c.do(http.MethodPut, "/api/nextcloud/settings", obj{"url": up.URL, "username": "anna", "appPassword": "wrong"}, http.StatusUnprocessableEntity)
	c.json(http.MethodPut, "/api/nextcloud/settings", obj{"url": up.URL, "username": "anna", "appPassword": "app-pw"}, http.StatusOK, nil)
	var nc obj
	c.json(http.MethodGet, "/api/nextcloud/status", nil, http.StatusOK, &nc)
	if nc["version"] != "30.0.2" || nc["server"] != nil || nc["serverHint"] == "" || nc["quota"].(obj)["percent"] != float64(25) {
		t.Fatalf("nextcloud: %v", nc)
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
	c.do(http.MethodGet, "/api/adguard/stats?provider=pihole", nil, http.StatusNotFound)
	c.do(http.MethodGet, "/api/adguard/stats?provider=bind", nil, http.StatusUnprocessableEntity)
	c.json(http.MethodPut, "/api/adguard/settings?provider=pihole", obj{"url": up.URL + "/pihole"}, http.StatusOK, nil)
	c.json(http.MethodGet, "/api/adguard/stats?provider=pihole", nil, http.StatusOK, &stats)
	if stats["blockedPercent"] != float64(10) || stats["avgProcessingMs"] != nil {
		t.Fatalf("pi-hole: %v", stats)
	}

	// --- proxmox (#84): read-only, actions off without PROXMOX_ACTIONS_GUESTS ---
	c.do(http.MethodGet, "/api/proxmox/status", nil, http.StatusNotFound)
	c.do(http.MethodPut, "/api/proxmox/settings", obj{"url": up.URL, "tokenId": "monitor@pve!th", "secret": "bad"}, http.StatusUnprocessableEntity)
	c.json(http.MethodPut, "/api/proxmox/settings", obj{"url": up.URL, "tokenId": "monitor@pve!th", "secret": "uuid"}, http.StatusOK, nil)
	if body := string(c.do(http.MethodGet, "/api/proxmox/settings", nil, http.StatusOK)); strings.Contains(body, "uuid") {
		t.Fatal("token secret leaked")
	}
	var pve obj
	c.json(http.MethodGet, "/api/proxmox/status", nil, http.StatusOK, &pve)
	if guests := pve["guests"].([]any); len(guests) != 1 || guests[0].(obj)["actions"] != false || pve["nodes"].([]any)[0].(obj)["cpuPercent"] != float64(50) {
		t.Fatalf("proxmox: %v", pve)
	}
	c.do(http.MethodPost, "/api/proxmox/guests/100/shutdown", nil, http.StatusForbidden)

	// --- smart home (#49): read-only without SMARTHOME_ACTIONS_ENTITIES ---
	c.do(http.MethodGet, "/api/smarthome/states?ids=switch.coffee", nil, http.StatusNotFound)
	c.json(http.MethodPut, "/api/smarthome/settings", obj{"url": up.URL + "/ha", "token": "ha-token"}, http.StatusOK, nil)
	if body := string(c.do(http.MethodGet, "/api/smarthome/settings", nil, http.StatusOK)); strings.Contains(body, "ha-token") {
		t.Fatal("home assistant token leaked")
	}
	var entities []obj
	c.json(http.MethodGet, "/api/smarthome/states?ids=switch.coffee", nil, http.StatusOK, &entities)
	if len(entities) != 1 || entities[0]["name"] != "Coffee" || entities[0]["action"] != "" {
		t.Fatalf("smarthome: %v", entities)
	}
	c.do(http.MethodPost, "/api/smarthome/entities/switch.coffee/action", nil, http.StatusForbidden)

	// --- *arr (#86): API keys stay on the server ---
	c.do(http.MethodGet, "/api/arr/overview", nil, http.StatusNotFound)
	c.do(http.MethodPost, "/api/arr/instances", obj{"app": "radarr", "url": up.URL + "/sonarr", "apiKey": "k"}, http.StatusUnprocessableEntity)
	c.json(http.MethodPost, "/api/arr/instances", obj{"app": "sonarr", "url": up.URL + "/sonarr", "apiKey": "sonarr-key"}, http.StatusCreated, nil)
	if body := string(c.do(http.MethodGet, "/api/arr/instances", nil, http.StatusOK)); strings.Contains(body, "sonarr-key") {
		t.Fatal("arr API key leaked")
	}
	var arrOverview obj
	c.json(http.MethodGet, "/api/arr/overview", nil, http.StatusOK, &arrOverview)
	if up := arrOverview["upcoming"].([]any); len(up) != 1 || arrOverview["missing"].(obj)["Sonarr"] != float64(2) {
		t.Fatalf("arr: %v", arrOverview)
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
	from := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)
	to := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	c.json(http.MethodGet, "/api/calendar/events?from="+from+"&to="+to, nil, http.StatusOK, &agenda)
	if len(agenda.Events) != 1 {
		t.Fatalf("calendar range: %+v", agenda)
	}
	c.do(http.MethodGet, "/api/calendar/events?from="+from+"&to="+time.Now().AddDate(0, 3, 0).UTC().Format(time.RFC3339), nil, http.StatusUnprocessableEntity)
	newEvent := obj{"calendarUrl": up.URL + "/dav/calendars/me/personal/", "summary": "Haircut", "allDay": false,
		"start": time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339), "end": time.Now().Add(25 * time.Hour).UTC().Format(time.RFC3339)}
	c.json(http.MethodPost, "/api/calendar/events", newEvent, http.StatusCreated, nil)
	newEvent["calendarUrl"] = "https://elsewhere.example/cal/"
	c.do(http.MethodPost, "/api/calendar/events", newEvent, http.StatusUnprocessableEntity)

	// --- browser bookmark import (Netscape HTML) ---
	html := `<!DOCTYPE NETSCAPE-Bookmark-file-1><DL><p><DT><H3>Imported</H3><DL><p>` +
		`<DT><A HREF="https://imported.example/" TAGS="a,b">Imported link</A></DL><p><DT><A HREF="javascript:x">No</A></DL>`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/bookmarks/import", strings.NewReader(html))
	req.Header.Set("Content-Type", "text/html")
	req.Header.Set("X-CSRF-Token", c.csrf)
	res, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var imported obj
	_ = json.NewDecoder(res.Body).Decode(&imported)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK || imported["created"] != float64(1) || imported["skippedInvalid"] != float64(1) || imported["foldersCreated"] != float64(1) {
		t.Fatalf("bookmark import: %d %v", res.StatusCode, imported)
	}

	// --- security log: the setup counts as the first sign-in of a new device ---
	var events []obj
	c.json(http.MethodGet, "/api/auth/events", nil, http.StatusOK, &events)
	seen := map[any]bool{}
	for _, e := range events {
		seen[e["type"]] = true
	}
	if !seen["login_succeeded"] || !seen["new_device"] {
		t.Fatalf("security log: %v", events)
	}

	// --- timeline (#74) ---
	var tl []obj
	c.json(http.MethodGet, "/api/timeline?limit=5", nil, http.StatusOK, &tl)
	c.do(http.MethodPut, "/api/timeline/settings", obj{"retentionDays": 0}, http.StatusUnprocessableEntity)
	c.do(http.MethodPut, "/api/timeline/settings", obj{"retentionDays": 14}, http.StatusOK)

	// --- disk forecast settings (#108) ---
	c.do(http.MethodPut, "/api/system/forecast-settings", obj{"windowDays": 2, "alertDays": 14}, http.StatusUnprocessableEntity)
	c.do(http.MethodPut, "/api/system/forecast-settings", obj{"windowDays": 30, "alertDays": 7}, http.StatusOK)

	// --- docker image updates (#107): off by default ---
	c.do(http.MethodGet, "/api/docker/updates", nil, http.StatusNotFound)
	c.do(http.MethodPut, "/api/docker/updates/settings", obj{"enabled": false, "registries": []obj{{"host": "ghcr.io", "username": "me"}}}, http.StatusUnprocessableEntity)
	var us obj
	c.json(http.MethodPut, "/api/docker/updates/settings", obj{"enabled": true, "semverTags": true,
		"registries": []obj{{"host": "ghcr.io", "username": "me", "password": "token"}}}, http.StatusOK, &us)
	if regs, _ := us["registries"].([]any); len(regs) != 1 || strings.Contains(fmt.Sprint(us), "token") {
		t.Fatalf("update settings: %+v", us)
	}
	c.do(http.MethodPut, "/api/docker/updates/settings", obj{"enabled": false}, http.StatusOK)

	// --- speedtest (#60): off until configured ---
	c.do(http.MethodGet, "/api/speedtest/results", nil, http.StatusNotFound)
	var st obj
	c.json(http.MethodGet, "/api/speedtest/settings", nil, http.StatusOK, &st)
	if st["configured"] != false {
		t.Fatalf("speedtest settings: %+v", st)
	}

	// --- disk health (#59): not set up means 404 with a hint ---
	c.do(http.MethodGet, "/api/diskhealth", nil, http.StatusNotFound)

	// --- version and update check (off by default, no outbound request) ---
	var ver obj
	c.json(http.MethodGet, "/api/system/version", nil, http.StatusOK, &ver)
	if ver["checkEnabled"] != false || ver["latest"] != nil || ver["current"] == "" {
		t.Fatalf("version: %v", ver)
	}

	// --- export / import ---
	var export struct {
		Format   string                     `json:"format"`
		Version  int                        `json:"version"`
		Sections map[string]json.RawMessage `json:"sections"`
	}
	c.json(http.MethodGet, "/api/transfer/export", nil, http.StatusOK, &export)
	if export.Format != "trisse-home-export" || export.Sections["notes"] == nil || export.Sections["dashboard"] == nil {
		t.Fatalf("export: %v", export)
	}
	if strings.Contains(string(c.do(http.MethodGet, "/api/transfer/export", nil, http.StatusOK)), "e2e-password") {
		t.Fatal("export contains a secret")
	}
	var notesBefore []obj
	c.json(http.MethodGet, "/api/notes", nil, http.StatusOK, &notesBefore)
	c.json(http.MethodPost, "/api/notes", obj{"title": "added after export", "body": "x"}, http.StatusCreated, nil)
	var result struct {
		Imported []string          `json:"imported"`
		Failed   map[string]string `json:"failed"`
	}
	c.json(http.MethodPost, "/api/transfer/import", export, http.StatusOK, &result)
	if len(result.Failed) != 0 || len(result.Imported) != len(export.Sections) {
		t.Fatalf("import: %+v", result)
	}
	var notesAfter []obj
	c.json(http.MethodGet, "/api/notes", nil, http.StatusOK, &notesAfter)
	if len(notesAfter) != len(notesBefore) {
		t.Fatalf("import did not restore notes: %d vs %d", len(notesAfter), len(notesBefore))
	}
	c.do(http.MethodPost, "/api/transfer/import", obj{"format": "other", "version": 1}, http.StatusUnprocessableEntity)

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
