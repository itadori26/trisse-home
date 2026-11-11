package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// TestAPITokens uses a token like an automation would: no cookies, no CSRF
// token, only the Authorization header.
func TestAPITokens(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)
	var created obj
	owner.json(http.MethodPost, "/api/auth/tokens", obj{"name": "Shortcuts", "scopes": []string{"todos:write"}, "expiresInDays": 0}, http.StatusCreated, &created)
	token := created["token"].(string)

	call := func(method, path, tok string, body any) int {
		var b bytes.Buffer
		if body != nil {
			_ = json.NewEncoder(&b).Encode(body)
		}
		req, _ := http.NewRequest(method, srv.URL+path, &b)
		req.Header.Set("Content-Type", "application/json")
		if tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		return res.StatusCode
	}
	for _, c := range []struct {
		method, path, tok string
		body              any
		want              int
	}{
		{http.MethodPost, "/api/todos", token, obj{"title": "Milch kaufen"}, http.StatusCreated},
		{http.MethodGet, "/api/todos", token, nil, http.StatusOK},
		{http.MethodGet, "/api/notes", token, nil, http.StatusForbidden},
		{http.MethodGet, "/api/auth/tokens", token, nil, http.StatusForbidden},
		{http.MethodPost, "/api/auth/tokens", token, obj{"name": "x", "scopes": []string{"todos:read"}}, http.StatusForbidden},
		{http.MethodGet, "/api/todos", "th_invalid", nil, http.StatusUnauthorized},
		{http.MethodGet, "/api/todos", "", nil, http.StatusUnauthorized},
	} {
		if got := call(c.method, c.path, c.tok, c.body); got != c.want {
			t.Errorf("%s %s: %d, want %d", c.method, c.path, got, c.want)
		}
	}
	var list []obj
	owner.json(http.MethodGet, "/api/auth/tokens", nil, http.StatusOK, &list)
	if len(list) != 1 || list[0]["lastUsedAt"] == nil || list[0]["token"] != nil {
		t.Fatalf("token list: %v", list)
	}
	owner.do(http.MethodDelete, "/api/auth/tokens/"+created["id"].(string), nil, http.StatusNoContent)
	if got := call(http.MethodGet, "/api/todos", token, nil); got != http.StatusUnauthorized {
		t.Fatalf("revoked token: %d", got)
	}
}
