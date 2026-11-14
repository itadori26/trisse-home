package main

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"trisse-home/internal/platform/config"
)

var deviceDir string

func init() {
	registerModuleE2E(moduleE2E{
		Name: "device",
		Config: func(cfg *config.Config, _ string) {
			deviceDir, _ = os.MkdirTemp("", "trisse-device-")
			cfg.InstallMethod = "appliance"
			cfg.DeviceDir = deviceDir
			cfg.DisabledModules = slices.DeleteFunc(cfg.DisabledModules, func(m string) bool { return m == "device" })
		},
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			defer os.RemoveAll(deviceDir)
			var s obj
			c.json(http.MethodGet, "/api/device", nil, http.StatusOK, &s)
			if s["requested"] != "" || s["phase"] != "idle" {
				t.Fatalf("device: %v", s)
			}
			c.do(http.MethodPut, "/api/device/display-mode", obj{"mode": "desktop"}, http.StatusUnprocessableEntity)
			c.json(http.MethodPut, "/api/device/display-mode", obj{"mode": "kiosk"}, http.StatusOK, &s)
			if s["requested"] != "kiosk" || s["phase"] != "pending" {
				t.Fatalf("after request: %v", s)
			}
			req, _ := os.ReadFile(filepath.Join(deviceDir, "display-mode"))
			url, _ := os.ReadFile(filepath.Join(deviceDir, "kiosk-url"))
			if strings.TrimSpace(string(req)) != "kiosk" || !strings.HasPrefix(string(url), "http://localhost/kiosk#") {
				t.Fatalf("hand-over files: %q %q", req, url)
			}
			// The display's link works from loopback only; the test server listens on 127.0.0.1.
			k := newClient(t, c.base)
			token := strings.TrimSpace(strings.TrimPrefix(string(url), "http://localhost/kiosk#"))
			k.do(http.MethodPost, "/api/auth/kiosk/login", obj{"token": token}, http.StatusOK)
			var list []obj
			c.json(http.MethodGet, "/api/auth/kiosks", nil, http.StatusOK, &list)
			if len(list) != 1 || list[0]["name"] != "this-device" {
				t.Fatalf("device kiosk: %v", list)
			}
			c.json(http.MethodPut, "/api/device/display-mode", obj{"mode": "headless"}, http.StatusOK, &s)
			if _, err := os.Stat(filepath.Join(deviceDir, "kiosk-url")); !os.IsNotExist(err) {
				t.Fatal("the kiosk address must go with headless")
			}
			c.json(http.MethodGet, "/api/auth/kiosks", nil, http.StatusOK, &list)
			if len(list) != 0 {
				t.Fatalf("device kiosk not removed: %v", list)
			}
		},
	})
}
