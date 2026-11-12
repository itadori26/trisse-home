#!/bin/bash
# Boots the built image's root file system in a systemd-nspawn container and
# checks what a user sees after the first boot (ADR 0015):
#
#   sudo deploy/raspberrypi/test-image.sh <image.img>
#
# The image stays unchanged (read-only mount, writes go to a tmpfs overlay).
# Hardware-specific services (cloud-init, NetworkManager, resizing, swap) are
# masked; the container shares the host network, so the service is reachable
# on http://127.0.0.1/. Needs systemd-container (systemd-nspawn, machinectl).
set -euo pipefail

image=$(realpath "${1:?usage: test-image.sh <image.img>}")
machine=trisse-home-test
mnt=$(mktemp -d)
loop=""

cleanup() {
	set +e
	machinectl terminate "$machine" >/dev/null 2>&1
	sleep 2
	mountpoint -q "$mnt" && umount "$mnt"
	[ -n "$loop" ] && losetup -d "$loop"
	rmdir "$mnt"
}
trap cleanup EXIT

fail() {
	echo "FAIL: $*" >&2
	echo "--- failed units" >&2
	systemd-run -M "$machine" --wait --pipe --quiet systemctl --failed --no-pager >&2 || true
	echo "--- trisse-home log" >&2
	systemd-run -M "$machine" --wait --pipe --quiet journalctl -u trisse-home -u trisse-home-setup-info --no-pager -n 100 >&2 || true
	exit 1
}

# in_guest CMD...: runs a command inside the booted container.
in_guest() { systemd-run -M "$machine" --wait --pipe --quiet "$@"; }

loop=$(losetup --find --show --partscan --read-only "$image")
mount -o ro "${loop}p2" "$mnt"

masks=(cloud-init-local cloud-init-network cloud-init-main cloud-init cloud-config cloud-final
	NetworkManager NetworkManager-wait-online systemd-networkd-wait-online
	rpi-resize rpi-resize-swap-file rpi-zram-writeback dphys-swapfile systemd-zram-setup@zram0
	regenerate_ssh_host_keys sshswitch userconfig)
args=()
for unit in "${masks[@]}"; do args+=("systemd.mask=$unit.service"); done

echo "== boot"
systemd-nspawn --quiet --directory="$mnt" --volatile=overlay --machine="$machine" \
	--boot -- "${args[@]}" >"${NSPAWN_LOG:-/tmp/nspawn.log}" 2>&1 &

for _ in $(seq 90); do
	curl -fsS http://127.0.0.1/healthz >/dev/null 2>&1 && break
	sleep 2
done
curl -fsS http://127.0.0.1/healthz || fail "trisse-home not reachable on port 80"
echo

echo "== service"
[ "$(in_guest systemctl is-active trisse-home)" = active ] || fail "service not active"
[ "$(in_guest hostname)" = trisse-home ] || fail "hostname is not trisse-home"
in_guest systemctl show trisse-home -p MainPID -p NRestarts -p WatchdogUSec
uid=$(in_guest sh -c 'stat -c %U /proc/$(systemctl show -p MainPID --value trisse-home)')
[ "$uid" = trisse-home ] || fail "service runs as $uid"

echo "== secrets generated on this device"
in_guest sh -c 'grep -Eq "^SECRET_KEY=.{40,}" /etc/trisse-home/secret.env' || fail "no SECRET_KEY generated"
token=$(in_guest sed -n 's/^SETUP_TOKEN=//p' /etc/trisse-home/secret.env)
[ -n "$token" ] || fail "no setup token generated"
in_guest trisse-home-setup-token | tee /dev/stderr | grep -q "http://trisse-home.local " || fail "address shown without port 80"

echo "== setup token on the console"
for _ in $(seq 30); do
	in_guest test -f /etc/issue.d/trisse-home.issue && break
	sleep 2
done
in_guest grep -q "$token" /etc/issue.d/trisse-home.issue || fail "token not shown on the console"

echo "== first setup"
curl -fsS -X POST -H 'Content-Type: application/json' \
	-d "{\"setupToken\":\"$token\",\"password\":\"correct horse battery staple\"}" \
	http://127.0.0.1/api/auth/setup >/dev/null || fail "setup with the token failed"
in_guest systemctl start trisse-home-setup-info.service
in_guest test ! -e /etc/issue.d/trisse-home.issue || fail "token still shown after the setup"

echo "== memory after setup"
in_guest systemctl show trisse-home -p MemoryCurrent

echo "OK: image boots and works"
