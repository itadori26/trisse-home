package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// OPML import and export of the feed reader (#97). The import only stores
// the feeds; nothing is fetched until the next refresh.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "rss-opml",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			opml := `<?xml version="1.0"?><opml version="2.0"><head><title>x</title></head><body>
				<outline text="Tech"><outline type="rss" text="Go" xmlUrl="https://go.dev/blog/feed.atom"/></outline>
				<outline type="rss" text="Bad" xmlUrl="ftp://x/feed"/>
			</body></opml>`
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, c.base+"/api/rss/opml", strings.NewReader(opml))
			req.Header.Set("Content-Type", "text/x-opml")
			req.Header.Set("X-CSRF-Token", c.csrf)
			res, err := c.http.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(res.Body)
			_ = res.Body.Close()
			var report obj
			_ = json.Unmarshal(body, &report)
			if res.StatusCode != http.StatusOK || report["added"] != 1.0 || len(report["invalid"].([]any)) != 1 {
				t.Fatalf("import: %d %s", res.StatusCode, body)
			}
			var feeds []obj
			c.json(http.MethodGet, "/api/rss/feeds", nil, http.StatusOK, &feeds)
			found := false
			for _, f := range feeds {
				if f["url"] == "https://go.dev/blog/feed.atom" {
					found = f["category"] == "Tech" && f["title"] == "Go"
					c.do(http.MethodDelete, "/api/rss/feeds/"+f["id"].(string), nil, http.StatusNoContent)
				}
			}
			if !found {
				t.Fatalf("imported feed missing: %v", feeds)
			}
			out := string(c.do(http.MethodGet, "/api/rss/opml", nil, http.StatusOK))
			if !strings.Contains(out, `<opml version="2.0">`) {
				t.Fatalf("export: %s", out)
			}
		},
	})
}
