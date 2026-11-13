#!/bin/sh
# Adds .deb packages to the trisse-home APT repository and re-signs it
# (ADR 0015). The repository is a static directory, served by GitHub Pages:
#
#   APT_SIGNING_KEY_ID=<fingerprint> scripts/apt-repo.sh <repo-dir> <package.deb>...
#
# Layout (suite "stable", component "main", amd64 and arm64):
#   pool/main/t/trisse-home/*.deb
#   dists/stable/{InRelease,Release,Release.gpg}
#   dists/stable/main/binary-<arch>/Packages(.gz)
#   trisse-home-archive-keyring.asc    public key for Signed-By
#
# The newest APT_KEEP versions (default 5) stay in the pool, so that users can
# roll back with apt install trisse-home=<version>. Needs dpkg-dev and gpg with
# the signing key in the current keyring.
set -eu

repo=${1:?usage: apt-repo.sh <repo-dir> <package.deb>...}
shift
key=${APT_SIGNING_KEY_ID:?APT_SIGNING_KEY_ID missing}
keep=${APT_KEEP:-5}
suite=stable
archs="amd64 arm64"
pool="pool/main/t/trisse-home"

mkdir -p "$repo/$pool"
for deb in "$@"; do
	# Canonical name <package>_<version>_<arch>.deb: dpkg-scanpackages --arch
	# only picks up files named like that.
	name="$(dpkg-deb -f "$deb" Package)_$(dpkg-deb -f "$deb" Version)_$(dpkg-deb -f "$deb" Architecture).deb"
	cp "$deb" "$repo/$pool/$name"
done

cd "$repo"

# Keep the newest $keep versions (all architectures of a version together).
versions=$(for f in "$pool"/*.deb; do dpkg-deb -f "$f" Version; done | sort -u | sort -V -r)
echo "$versions" | tail -n +"$((keep + 1))" | while read -r old; do
	[ -n "$old" ] || continue
	echo "apt-repo: removing $old from the pool"
	for f in "$pool"/*.deb; do
		if [ "$(dpkg-deb -f "$f" Version)" = "$old" ]; then
			rm -f "$f"
		fi
	done
done

dist="dists/$suite"
rm -rf "$dist"
for arch in $archs; do
	dir="$dist/main/binary-$arch"
	mkdir -p "$dir"
	dpkg-scanpackages --multiversion --arch "$arch" pool >"$dir/Packages" 2>/dev/null
	gzip -9nkf "$dir/Packages"
done

# hashes <command>: "<hash> <size> <path>" for every index file.
hashes() {
	(cd "$dist" && find main -type f -name 'Packages*' | sort | while read -r f; do
		printf ' %s %s %s\n' "$($1 "$f" | cut -d' ' -f1)" "$(stat -c %s "$f")" "$f"
	done)
}

{
	echo "Origin: trisse-home"
	echo "Label: trisse-home"
	echo "Suite: $suite"
	echo "Codename: $suite"
	echo "Date: $(LC_ALL=C date -u '+%a, %d %b %Y %H:%M:%S UTC')"
	echo "Architectures: $archs"
	echo "Components: main"
	echo "Description: trisse-home self-hosted start page"
	echo "MD5Sum:"
	hashes md5sum
	echo "SHA256:"
	hashes sha256sum
} >"$dist/Release"

gpg --batch --yes --local-user "$key" --clearsign --output "$dist/InRelease" "$dist/Release"
gpg --batch --yes --local-user "$key" --armor --detach-sign --output "$dist/Release.gpg" "$dist/Release"
gpg --batch --yes --armor --export "$key" >trisse-home-archive-keyring.asc

echo "apt-repo: $(ls "$pool" | wc -l) packages, versions:" $(for f in "$pool"/*.deb; do dpkg-deb -f "$f" Version; done | sort -u | sort -V -r)
