package domain

import "testing"

func TestParseImageRef(t *testing.T) {
	tests := []struct {
		in                       string
		registry, repo, tag, dig string
		canonical                string
	}{
		{"nginx", DockerHub, "library/nginx", "latest", "", "nginx:latest"},
		{"nginx:1.27", DockerHub, "library/nginx", "1.27", "", "nginx:1.27"},
		{"linuxserver/wireguard:latest", DockerHub, "linuxserver/wireguard", "latest", "", "linuxserver/wireguard:latest"},
		{"ghcr.io/itadori26/trisse-home:v1.2.0", "ghcr.io", "itadori26/trisse-home", "v1.2.0", "", "ghcr.io/itadori26/trisse-home:v1.2.0"},
		{"registry.local:5000/app", "registry.local:5000", "app", "latest", "", "registry.local:5000/app:latest"},
		{"localhost/app:dev", "localhost", "app", "dev", "", "localhost/app:dev"},
		{"redis@sha256:abc", DockerHub, "library/redis", "", "sha256:abc", "redis@sha256:abc"},
		{"index.docker.io/library/postgres:16", DockerHub, "library/postgres", "16", "", "postgres:16"},
	}
	for _, tc := range tests {
		r, ok := ParseImageRef(tc.in)
		if !ok || r.Registry != tc.registry || r.Repository != tc.repo || r.Tag != tc.tag || r.Digest != tc.dig || r.String() != tc.canonical {
			t.Errorf("%q: %+v %v (%s)", tc.in, r, ok, r.String())
		}
	}
	if _, ok := ParseImageRef("sha256:0123"); ok {
		t.Error("image IDs are not references")
	}
	if r, _ := ParseImageRef("nginx"); r.APIHost() != "registry-1.docker.io" || r.Pinned() {
		t.Error("docker hub api host")
	}
}

func TestCompareDigest(t *testing.T) {
	local := []string{"nginx@sha256:aaa"}
	if CompareDigest(local, "sha256:aaa") != UpToDate || CompareDigest(local, "sha256:bbb") != UpdateAvailable || CompareDigest(nil, "sha256:aaa") != UpdateLocal {
		t.Fatal("compare")
	}
}

func TestNewerTag(t *testing.T) {
	tags := []string{"1.26.0", "1.27.1", "1.27.10", "1.28-rc1", "v1.30.0", "1.29", "latest", "2.0.0"}
	if got := NewerTag("1.27.1", tags); got != "2.0.0" {
		t.Fatalf("got %q", got)
	}
	if got := NewerTag("1.28", tags); got != "1.29" {
		t.Fatalf("two-part: %q", got)
	}
	if got := NewerTag("2.0.0", tags); got != "" {
		t.Fatalf("newest: %q", got)
	}
	if NewerTag("latest", tags) != "" || !IsSemverTag("v1.2.3") || IsSemverTag("1.2.3-alpine") || IsSemverTag("01.2") {
		t.Fatal("semver detection")
	}
}

func TestReleaseNotesURL(t *testing.T) {
	cases := map[string]string{
		"https://github.com/itadori26/trisse-home":     "https://github.com/itadori26/trisse-home/releases",
		"https://github.com/itadori26/trisse-home.git": "https://github.com/itadori26/trisse-home/releases",
		"https://gitlab.com/group/project":             "https://gitlab.com/group/project",
		"javascript:alert(1)":                          "",
		"":                                             "",
	}
	for in, want := range cases {
		if got := ReleaseNotesURL(map[string]string{SourceLabel: in}); got != want {
			t.Errorf("%q: %q, want %q", in, got, want)
		}
	}
}
