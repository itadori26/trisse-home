package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/containers/json" || r.URL.Query().Get("all") != "1" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"Id":"0123456789abcdef","Names":["/homarr"],"Image":"ghcr.io/x:latest","State":"running",
			"Status":"Up 3 hours (healthy)","Created":1700000000,"Labels":{"com.docker.compose.project":"homarr"}}]`))
	}))
	defer srv.Close()
	got, err := NewAPI(srv.Client(), srv.URL+"/").Containers(context.Background())
	if err != nil || len(got) != 1 {
		t.Fatalf("unexpected %v %v", got, err)
	}
	c := got[0]
	if c.ID != "0123456789ab" || c.Name != "homarr" || c.Project != "homarr" || c.Created.Year() != 2023 {
		t.Fatalf("mapping wrong: %+v", c)
	}
	if _, err := NewAPI(srv.Client(), srv.URL+"/nope").Containers(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
