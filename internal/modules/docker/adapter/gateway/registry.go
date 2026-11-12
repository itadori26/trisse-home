package gateway

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"trisse-home/internal/modules/docker/domain"
	"trisse-home/internal/modules/docker/usecase"
)

// manifestTypes are accepted so the registry returns the digest of the
// multi-arch index, which is what `docker pull` records in RepoDigests.
var manifestTypes = strings.Join([]string{
	"application/vnd.oci.image.index.v1+json",
	"application/vnd.docker.distribution.manifest.list.v2+json",
	"application/vnd.oci.image.manifest.v1+json",
	"application/vnd.docker.distribution.manifest.v2+json",
}, ", ")

const maxTags = 2000

// Registry implements usecase.RegistryClient for the OCI distribution API
// with anonymous or basic-auth bearer tokens (Docker Hub, GHCR, Quay,
// private registries).
type Registry struct {
	client *http.Client
	// scheme is https; tests use http.
	scheme string
}

// NewRegistry returns a registry client.
func NewRegistry(client *http.Client) *Registry { return &Registry{client: client, scheme: "https"} }

// Digest implements usecase.RegistryClient with a HEAD request (not
// counted against Docker Hub's pull rate limit).
func (r *Registry) Digest(ctx context.Context, ref domain.ImageRef, cred *usecase.RegistryCredential) (string, error) {
	u := fmt.Sprintf("%s://%s/v2/%s/manifests/%s", r.scheme, ref.APIHost(), ref.Repository, url.PathEscape(ref.Tag))
	res, err := r.do(ctx, http.MethodHead, u, ref, cred, manifestTypes)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	d := res.Header.Get("Docker-Content-Digest")
	if d == "" {
		return "", errors.New("the registry sent no digest")
	}
	return d, nil
}

// Tags implements usecase.RegistryClient (first page, up to maxTags).
func (r *Registry) Tags(ctx context.Context, ref domain.ImageRef, cred *usecase.RegistryCredential) ([]string, error) {
	u := fmt.Sprintf("%s://%s/v2/%s/tags/list?n=%d", r.scheme, ref.APIHost(), ref.Repository, maxTags)
	res, err := r.do(ctx, http.MethodGet, u, ref, cred, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var payload struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode tags: %w", err)
	}
	return payload.Tags, nil
}

// do sends the request and, on 401, fetches a token as described by the
// WWW-Authenticate challenge and retries once.
func (r *Registry) do(ctx context.Context, method, u string, ref domain.ImageRef, cred *usecase.RegistryCredential, accept string) (*http.Response, error) {
	send := func(auth string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, method, u, nil)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Accept", accept)
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		return r.client.Do(req)
	}
	res, err := send("")
	if err != nil {
		return nil, fmt.Errorf("registry request: %w", err)
	}
	if res.StatusCode == http.StatusUnauthorized {
		challenge := res.Header.Get("WWW-Authenticate")
		_ = res.Body.Close()
		auth, err := r.authorize(ctx, challenge, ref, cred)
		if err != nil {
			return nil, err
		}
		if res, err = send(auth); err != nil {
			return nil, fmt.Errorf("registry request: %w", err)
		}
	}
	if res.StatusCode != http.StatusOK {
		_ = res.Body.Close()
		return nil, fmt.Errorf("registry answered %s", res.Status)
	}
	return res, nil
}

func (r *Registry) authorize(ctx context.Context, challenge string, ref domain.ImageRef, cred *usecase.RegistryCredential) (string, error) {
	scheme, params := parseChallenge(challenge)
	switch strings.ToLower(scheme) {
	case "basic":
		if cred == nil {
			return "", errors.New("the registry requires credentials")
		}
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(cred.Username+":"+cred.Password)), nil
	case "bearer":
	default:
		return "", fmt.Errorf("unsupported registry authentication %q", scheme)
	}
	realm, err := url.Parse(params["realm"])
	if err != nil || (realm.Scheme != "https" && realm.Scheme != r.scheme) || realm.Host == "" {
		return "", errors.New("invalid token realm")
	}
	q := realm.Query()
	if s := params["service"]; s != "" {
		q.Set("service", s)
	}
	scope := params["scope"]
	if scope == "" {
		scope = "repository:" + ref.Repository + ":pull"
	}
	q.Set("scope", scope)
	realm.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, realm.String(), nil)
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	if cred != nil {
		req.SetBasicAuth(cred.Username, cred.Password)
	}
	res, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint answered %s", res.Status)
	}
	var tok struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&tok); err != nil {
		return "", fmt.Errorf("decode token: %w", err)
	}
	t := tok.Token
	if t == "" {
		t = tok.AccessToken
	}
	if t == "" {
		return "", errors.New("the token endpoint sent no token")
	}
	return "Bearer " + t, nil
}

// parseChallenge splits `Bearer realm="…",service="…",scope="…"`.
func parseChallenge(h string) (string, map[string]string) {
	scheme, rest, _ := strings.Cut(strings.TrimSpace(h), " ")
	params := map[string]string{}
	for rest != "" {
		var kv string
		rest = strings.TrimLeft(rest, " ,")
		key, after, ok := strings.Cut(rest, "=")
		if !ok {
			break
		}
		if strings.HasPrefix(after, `"`) {
			end := strings.Index(after[1:], `"`)
			if end < 0 {
				break
			}
			kv, rest = after[1:1+end], after[2+end:]
		} else {
			kv, rest, _ = strings.Cut(after, ",")
		}
		params[strings.ToLower(strings.TrimSpace(key))] = kv
	}
	return scheme, params
}
