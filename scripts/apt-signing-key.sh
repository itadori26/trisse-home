#!/bin/sh
# One-time setup of the APT repository's signing key, run by a maintainer
# (ADR 0015, docs/native-install.md):
#
#   scripts/apt-signing-key.sh
#
# 1. creates an Ed25519 signing key in a temporary GnuPG home,
# 2. stores the private key as the Actions secret APT_SIGNING_KEY (gh CLI),
# 3. writes the public key to packaging/apt/trisse-home-archive-keyring.asc.
#
# The private key exists only in the temporary directory, which is deleted at
# the end, and in the GitHub secret. Commit the public key afterwards: the
# Raspberry Pi image only configures the repository once it is in the tree.
# Rotation: run again, commit the new public key, release; devices pick up
# the new key with the next image or by updating the keyring file.
set -eu

repo_root=$(git rev-parse --show-toplevel)
public="$repo_root/packaging/apt/trisse-home-archive-keyring.asc"

command -v gh >/dev/null || { echo "needs the GitHub CLI (gh)" >&2; exit 1; }
gh auth status >/dev/null

GNUPGHOME=$(mktemp -d)
export GNUPGHOME
trap 'rm -rf "$GNUPGHOME"' EXIT

gpg --batch --quiet --passphrase '' \
	--quick-gen-key "trisse-home APT repository <noreply@github.com>" ed25519 sign 5y
fpr=$(gpg --list-keys --with-colons | awk -F: '/^fpr/ { print $10; exit }')

gpg --batch --armor --export-secret-keys "$fpr" | gh secret set APT_SIGNING_KEY
mkdir -p "$(dirname "$public")"
gpg --batch --armor --export "$fpr" >"$public"

echo "Signing key $fpr"
echo "  private key -> Actions secret APT_SIGNING_KEY"
echo "  public key  -> ${public#"$repo_root"/} (commit it)"
echo "Then enable GitHub Pages: Settings -> Pages -> Deploy from branch gh-pages, / (root)."
