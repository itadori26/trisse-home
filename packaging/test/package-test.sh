#!/bin/sh
# Install/upgrade/remove/purge test for the trisse-home .deb.
#
# Run as root in a throwaway Debian container without a running systemd, e.g.:
#   docker run --rm -v "$PWD:/src:ro" -w /src debian:trixie \
#     sh packaging/test/package-test.sh dist/trisse-home_*_amd64.deb
# The service itself is started and probed separately on a systemd host (CI).
set -eu

deb=${1:?usage: package-test.sh <package.deb>}
conf=/etc/trisse-home
state=/var/lib/trisse-home

step() { printf '\n== %s\n' "$*"; }
fail() {
	echo "FAIL: $*" >&2
	exit 1
}
mode_of() { stat -c '%U:%G %a' "$1"; }
key_of() { sed -n 's/^SECRET_KEY=//p' "$1"; }
token_of() { sed -n 's/^SETUP_TOKEN=//p' "$1"; }

export DEBIAN_FRONTEND=noninteractive

step "install"
apt-get update -qq
apt-get install -y -qq "$(realpath "$deb")" >/dev/null
getent passwd trisse-home >/dev/null || fail "system user missing"
[ "$(getent passwd trisse-home | cut -d: -f7)" = /usr/sbin/nologin ] || fail "user must not have a login shell"
[ "$(mode_of $conf)" = "root:trisse-home 750" ] || fail "$conf: $(mode_of $conf)"
[ "$(mode_of $conf/trisse-home.env)" = "root:trisse-home 640" ] || fail "config: $(mode_of $conf/trisse-home.env)"
[ ! -e $conf/secret.env ] || fail "secret must not exist before the first start"
[ -x /usr/bin/trisse-home ] || fail "binary missing"
/usr/bin/trisse-home version
[ -f /usr/lib/systemd/system/trisse-home.service ] || fail "unit missing"
# Enabled for the next boot even without a running systemd (image builds).
[ -L /etc/systemd/system/multi-user.target.wants/trisse-home.service ] || fail "unit not enabled"

step "first start creates a per-installation key and setup token"
/usr/lib/trisse-home/trisse-home-init
[ "$(mode_of $conf/secret.env)" = "root:trisse-home 640" ] || fail "secret: $(mode_of $conf/secret.env)"
key=$(key_of $conf/secret.env)
[ "$(printf %s "$key" | base64 -d | wc -c)" -eq 32 ] || fail "SECRET_KEY is not 32 bytes base64"
token=$(token_of $conf/secret.env)
echo "$token" | grep -Eq '^[A-HJ-NP-Z2-9]{4}(-[A-HJ-NP-Z2-9]{4}){4}$' || fail "unexpected setup token format: $token"
/usr/lib/trisse-home/trisse-home-init
[ "$(key_of $conf/secret.env)" = "$key" ] || fail "second start must keep the key"
[ "$(token_of $conf/secret.env)" = "$token" ] || fail "second start must keep the setup token"
[ "$(grep -c '^SECRET_KEY=' $conf/secret.env)" -eq 1 ] || fail "duplicate SECRET_KEY lines"

step "setup token helper"
trisse-home-setup-token | grep -q "Token  $token" || fail "trisse-home-setup-token does not show the token"
[ -f /usr/lib/systemd/system/trisse-home-setup-info.timer ] || fail "setup info timer missing"
[ -L /etc/systemd/system/timers.target.wants/trisse-home-setup-info.timer ] || fail "setup info timer not enabled"

step "upgrade keeps local changes"
echo "LOG_LEVEL=debug" >>$conf/trisse-home.env
apt-get install -y -qq --reinstall -o Dpkg::Options::=--force-confold "$(realpath "$deb")" >/dev/null
grep -q '^LOG_LEVEL=debug' $conf/trisse-home.env || fail "local config change lost"
[ "$(key_of $conf/secret.env)" = "$key" ] || fail "upgrade changed the key"
[ "$(mode_of $conf/trisse-home.env)" = "root:trisse-home 640" ] || fail "upgrade changed config permissions"

step "remove keeps configuration and data"
mkdir -p $state && touch $state/trisse-home.db
apt-get remove -y -qq trisse-home >/dev/null
[ ! -e /usr/bin/trisse-home ] || fail "binary still installed"
[ -f $conf/trisse-home.env ] && [ -f $conf/secret.env ] || fail "remove deleted configuration"
[ -f $state/trisse-home.db ] || fail "remove deleted data"

step "purge deletes configuration, keeps data and the key next to it"
apt-get purge -y -qq trisse-home >/dev/null
[ ! -e $conf ] || fail "$conf still exists"
[ -f $state/trisse-home.db ] || fail "purge deleted data"
[ "$(mode_of $state/secret.env.purged)" = "root:root 600" ] || fail "saved key: $(mode_of $state/secret.env.purged 2>/dev/null || echo missing)"
[ ! -L /etc/systemd/system/multi-user.target.wants/trisse-home.service ] || fail "unit still enabled"

step "reinstall after purge restores the key"
apt-get install -y -qq "$(realpath "$deb")" >/dev/null
/usr/lib/trisse-home/trisse-home-init
[ "$(key_of $conf/secret.env)" = "$key" ] || fail "key not restored"
[ ! -e $state/secret.env.purged ] || fail "saved key not cleaned up"

step "an admin-provided key wins"
apt-get purge -y -qq trisse-home >/dev/null
rm -rf $state
apt-get install -y -qq "$(realpath "$deb")" >/dev/null
echo "SECRET_KEY=$key" >>$conf/trisse-home.env
echo "SETUP_TOKEN=my-own-setup-token-123" >>$conf/trisse-home.env
/usr/lib/trisse-home/trisse-home-init
! grep -q '^SECRET_KEY=' $conf/secret.env 2>/dev/null || fail "key generated although one is configured"
! grep -q '^SETUP_TOKEN=' $conf/secret.env 2>/dev/null || fail "token generated although one is configured"
trisse-home-setup-token | grep -q "Token  my-own-setup-token-123" || fail "configured token not shown"

printf '\nOK: all package tests passed\n'
