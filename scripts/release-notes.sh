#!/bin/sh
# Prints Markdown release notes for a tag from Conventional Commits since the
# previous tag. Usage: scripts/release-notes.sh v1.2.0
set -eu
tag=${1:?usage: $0 <tag>}
prev=$(git describe --tags --abbrev=0 --match 'v[0-9]*' "${tag}^" 2>/dev/null || true)
range=${prev:+$prev..}$tag

section() {
	title=$1
	pattern=$2
	lines=$(git log --no-merges --format='%s' "$range" | grep -E "$pattern" | sed -E 's/^[a-z]+(\(([^)]*)\))?!?: /- **\2** /; s/- \*\*\*\* /- /' || true)
	if [ -n "$lines" ]; then
		printf '### %s\n\n%s\n\n' "$title" "$lines"
	fi
}

printf '## %s\n\n' "$tag"
section 'Breaking changes' '^[a-z]+(\([^)]*\))?!:'
section 'Features' '^feat(\([^)]*\))?:'
section 'Fixes' '^fix(\([^)]*\))?:'
section 'Security' '^(sec|security)(\([^)]*\))?:'
section 'Performance' '^perf(\([^)]*\))?:'
section 'Documentation' '^docs(\([^)]*\))?:'
section 'Maintenance' '^(chore|build|ci|refactor|style|test)(\([^)]*\))?:'
if [ -n "$prev" ]; then
	printf '**Full changelog:** https://github.com/itadori26/trisse-home/compare/%s...%s\n' "$prev" "$tag"
fi
