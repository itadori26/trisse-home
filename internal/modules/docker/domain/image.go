package domain

import (
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

// UpdateCheckInterval is how often an image is asked at its registry at most
// (Docker Hub rate limits), unless its local digest changed (#107).
const UpdateCheckInterval = 6 * time.Hour

// DockerHub is the default registry of unqualified image names.
const DockerHub = "docker.io"

// ImageRef is a parsed image reference.
type ImageRef struct {
	Registry   string // e.g. docker.io, ghcr.io, registry.local:5000
	Repository string // e.g. library/nginx, linuxserver/wireguard
	Tag        string // defaults to latest
	Digest     string // set when the reference pins a digest
}

// ParseImageRef parses references such as "nginx", "nginx:1.27",
// "ghcr.io/owner/app:v1" or "repo@sha256:…"; ok=false for image IDs.
func ParseImageRef(s string) (ImageRef, bool) {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "sha256:") {
		return ImageRef{}, false
	}
	var r ImageRef
	if i := strings.Index(s, "@"); i >= 0 {
		r.Digest, s = s[i+1:], s[:i]
	}
	// A colon after the last slash separates the tag.
	if i := strings.LastIndex(s, ":"); i > strings.LastIndex(s, "/") {
		r.Tag, s = s[i+1:], s[:i]
	}
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 2 && (strings.ContainsAny(parts[0], ".:") || parts[0] == "localhost") {
		r.Registry, r.Repository = parts[0], parts[1]
	} else {
		r.Registry, r.Repository = DockerHub, s
	}
	if r.Registry == "index.docker.io" || r.Registry == "registry-1.docker.io" {
		r.Registry = DockerHub
	}
	if r.Registry == DockerHub && !strings.Contains(r.Repository, "/") {
		r.Repository = "library/" + r.Repository
	}
	if r.Tag == "" && r.Digest == "" {
		r.Tag = "latest"
	}
	if r.Repository == "" {
		return ImageRef{}, false
	}
	return r, true
}

// APIHost is the host serving the distribution API.
func (r ImageRef) APIHost() string {
	if r.Registry == DockerHub {
		return "registry-1.docker.io"
	}
	return r.Registry
}

// Pinned reports whether the reference fixes a digest.
func (r ImageRef) Pinned() bool { return r.Digest != "" }

// String renders the canonical reference.
func (r ImageRef) String() string {
	repo := r.Repository
	if r.Registry == DockerHub {
		repo = strings.TrimPrefix(repo, "library/")
	} else {
		repo = r.Registry + "/" + repo
	}
	if r.Digest != "" {
		return repo + "@" + r.Digest
	}
	return repo + ":" + r.Tag
}

// UpdateStatus is the result of an update check.
type UpdateStatus string

// Update statuses.
const (
	UpdateUnknown   UpdateStatus = "unknown"
	UpToDate        UpdateStatus = "up_to_date"
	UpdateAvailable UpdateStatus = "update_available"
	// UpdatePinned: the container uses a fixed digest; nothing to compare.
	UpdatePinned UpdateStatus = "pinned"
	// UpdateLocal: the image was built locally and has no registry digest.
	UpdateLocal UpdateStatus = "local"
	UpdateError UpdateStatus = "error"
)

// CompareDigest decides the status from the repo digests of the local
// image ("repo@sha256:…") and the digest the registry has for the tag.
func CompareDigest(repoDigests []string, remote string) UpdateStatus {
	if len(repoDigests) == 0 {
		return UpdateLocal
	}
	for _, d := range repoDigests {
		if d == remote || strings.HasSuffix(d, "@"+remote) {
			return UpToDate
		}
	}
	return UpdateAvailable
}

type semver struct {
	prefix string
	nums   []int
}

func parseSemver(tag string) (semver, bool) {
	v := semver{}
	if strings.HasPrefix(tag, "v") {
		v.prefix, tag = "v", tag[1:]
	}
	parts := strings.Split(tag, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return semver{}, false
	}
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || (len(p) > 1 && p[0] == '0') {
			return semver{}, false
		}
		v.nums = append(v.nums, n)
	}
	return v, true
}

// IsSemverTag reports whether a tag looks like 1.2, 1.2.3 or v1.2.3.
func IsSemverTag(tag string) bool {
	_, ok := parseSemver(tag)
	return ok
}

// NewerTag returns the highest tag of the same shape (prefix and number of
// parts) that is greater than current, or "".
func NewerTag(current string, tags []string) string {
	cur, ok := parseSemver(current)
	if !ok {
		return ""
	}
	best, bestTag := cur, ""
	for _, t := range tags {
		v, ok := parseSemver(t)
		if !ok || v.prefix != cur.prefix || len(v.nums) != len(cur.nums) {
			continue
		}
		if slices.Compare(v.nums, best.nums) > 0 {
			best, bestTag = v, t
		}
	}
	return bestTag
}

// SourceLabel is the OCI label pointing to the source repository.
const SourceLabel = "org.opencontainers.image.source"

// ReleaseNotesURL derives a release notes link from the source label.
func ReleaseNotesURL(labels map[string]string) string {
	src := strings.TrimSuffix(strings.TrimSpace(labels[SourceLabel]), ".git")
	u, err := url.Parse(src)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return ""
	}
	if u.Host == "github.com" && strings.Count(strings.Trim(u.Path, "/"), "/") == 1 {
		return "https://github.com" + strings.TrimRight(u.Path, "/") + "/releases"
	}
	return u.String()
}
