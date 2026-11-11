package domain

import (
	"strconv"
	"strings"
)

// Version is a semantic version such as 1.4.2. Builds that are not releases
// (e.g. "dev" or a commit hash) are not comparable.
type Version struct {
	raw                 string
	major, minor, patch int
	valid               bool
}

// ParseVersion accepts "1.2.3" and "v1.2.3"; pre-release and build suffixes
// ("1.2.3-rc.1+abc") keep the release numbers but mark it as pre-release.
func ParseVersion(s string) Version {
	v := Version{raw: strings.TrimSpace(s)}
	core := strings.TrimPrefix(v.raw, "v")
	if i := strings.IndexAny(core, "-+"); i >= 0 {
		core = core[:i]
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return v
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return v
		}
		nums[i] = n
	}
	v.major, v.minor, v.patch, v.valid = nums[0], nums[1], nums[2], true
	return v
}

// String returns the version as given.
func (v Version) String() string { return v.raw }

// IsRelease reports whether the version is a comparable release version.
func (v Version) IsRelease() bool { return v.valid }

// IsNewerThan reports whether v is a newer release than o. Non-release
// builds are never older or newer than anything, so a development build
// never shows an update hint.
func (v Version) IsNewerThan(o Version) bool {
	if !v.valid || !o.valid {
		return false
	}
	if v.major != o.major {
		return v.major > o.major
	}
	if v.minor != o.minor {
		return v.minor > o.minor
	}
	return v.patch > o.patch
}
