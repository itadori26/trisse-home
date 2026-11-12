#!/bin/bash
# Builds the trisse-home Raspberry Pi image (ADR 0015) by customising the
# official Raspberry Pi OS Lite image pinned in base.env:
#
#   sudo deploy/raspberrypi/build-image.sh <trisse-home_*_arm64.deb> <version> [out-dir]
#
# Runs as root on an arm64 Linux host (e.g. the GitHub runner ubuntu-24.04-arm,
# or a Raspberry Pi), so the image's own binaries run in a chroot without
# emulation. Needs: curl, xz, parted, e2fsprogs, util-linux (losetup).
#
# What it changes on top of Raspberry Pi OS Lite:
#   - installs the trisse-home package and unattended-upgrades
#   - files from rootfs/ (port 80, journal limits, automatic updates)
#   - hostname "trisse-home" (Imager customisation can still change it)
# Every device generates its own secrets on the first boot; the build checks
# that the image contains none.
set -euo pipefail

here=$(cd "$(dirname "$0")" && pwd)
deb=$(realpath "${1:?usage: build-image.sh <package.deb> <version> [out-dir]}")
version=${2:?version missing}
out=$(realpath -m "${3:-dist}")
work=${WORK_DIR:-$here/.work}
extra_mb=${IMAGE_EXTRA_MB:-512}

# shellcheck source=deploy/raspberrypi/base.env
. "$here/base.env"

name="trisse-home-$version-raspberrypi-arm64"
image="$work/$name.img"
root="$work/root"
loop=""

log() { printf '\n== %s\n' "$*"; }

cleanup() {
	set +e
	for m in boot/firmware dev/pts dev proc sys; do
		mountpoint -q "$root/$m" && umount -l "$root/$m"
	done
	mountpoint -q "$root" && umount -l "$root"
	[ -n "$loop" ] && losetup -d "$loop"
}
trap cleanup EXIT

[ "$(id -u)" -eq 0 ] || { echo "run as root" >&2; exit 1; }
[ "$(uname -m)" = aarch64 ] || { echo "run on an arm64 host (no emulation)" >&2; exit 1; }
mkdir -p "$work" "$out"

log "base image: $BASE_NAME"
base="$work/$(basename "$BASE_URL")"
if ! echo "$BASE_SHA256  $base" | sha256sum --check --status 2>/dev/null; then
	curl -fL --retry 3 -o "$base" "$BASE_URL"
fi
echo "$BASE_SHA256  $base" | sha256sum --check
xz -dc "$base" >"$image"

log "grow the root partition by $extra_mb MiB"
truncate -s "+${extra_mb}M" "$image"
parted -s "$image" resizepart 2 100%
loop=$(losetup --find --show --partscan "$image")
e2fsck -fy "${loop}p2" >/dev/null || [ $? -le 1 ]
resize2fs "${loop}p2" >/dev/null

log "mount"
mkdir -p "$root"
mount "${loop}p2" "$root"
mount "${loop}p1" "$root/boot/firmware"
mount --bind /dev "$root/dev"
mount --bind /dev/pts "$root/dev/pts"
mount -t proc proc "$root/proc"
mount -t sysfs sys "$root/sys"
# /run stays the image's own (empty) directory: the package must not try to
# start services in the chroot. policy-rc.d makes sure of it.
printf '#!/bin/sh\nexit 101\n' >"$root/usr/sbin/policy-rc.d"
chmod 755 "$root/usr/sbin/policy-rc.d"
resolv_backup=""
if [ -e "$root/etc/resolv.conf" ] || [ -L "$root/etc/resolv.conf" ]; then
	resolv_backup="$root/etc/resolv.conf.trisse-home-build"
	mv "$root/etc/resolv.conf" "$resolv_backup"
fi
cp -L /etc/resolv.conf "$root/etc/resolv.conf"

log "base image facts (for the build log)"
cat "$root/etc/os-release" | grep -E '^(PRETTY_NAME|VERSION_CODENAME)='
ls -1 "$root/boot/firmware" | grep -Ev '\.(dtbo?|elf|dat|bin)$' | tr '\n' ' '; echo
for f in user-data meta-data network-config; do
	[ -f "$root/boot/firmware/$f" ] && { echo "--- boot/firmware/$f"; sed -n '1,20p' "$root/boot/firmware/$f"; }
done

log "install packages"
cp "$deb" "$root/tmp/trisse-home.deb"
chroot "$root" /bin/sh -eu <<'EOF'
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq --no-install-recommends avahi-daemon unattended-upgrades /tmp/trisse-home.deb
apt-cache policy | grep -E '^ +release' | sort -u
EOF
rm -f "$root/tmp/trisse-home.deb"

log "appliance configuration"
# Without -a/-p, files copied by root belong to root (not to the CI user).
cp -r "$here/rootfs/." "$root/"
chroot "$root" chown root:trisse-home /etc/trisse-home/appliance.env
chmod 0640 "$root/etc/trisse-home/appliance.env"
old_host=$(cat "$root/etc/hostname")
echo trisse-home >"$root/etc/hostname"
sed -i "s/\b$old_host\b/trisse-home/g" "$root/etc/hosts"
# Root file system without access times (fewer writes on the SD card).
sed -i -E '/[[:space:]]\/[[:space:]]/ { /noatime/! s/(ext4[[:space:]]+)([^[:space:]]+)/\1\2,noatime/ }' "$root/etc/fstab"
grep -E '[[:space:]]/[[:space:]]' "$root/etc/fstab"
chroot "$root" systemctl is-enabled trisse-home.service trisse-home-setup-info.timer avahi-daemon.service

log "clean up"
chroot "$root" apt-get clean
rm -rf "$root"/var/lib/apt/lists/* "$root"/var/cache/apt/*.bin "$root"/var/log/apt/* \
	"$root"/var/log/dpkg.log "$root"/root/.bash_history
rm -f "$root/usr/sbin/policy-rc.d" "$root/etc/resolv.conf"
[ -n "$resolv_backup" ] && mv "$resolv_backup" "$root/etc/resolv.conf"

log "check: no secrets in the image"
fail=0
if [ -e "$root/etc/trisse-home/secret.env" ]; then echo "secret.env present"; fail=1; fi
if grep -Eqs '^[[:space:]]*(SECRET_KEY|SETUP_TOKEN)=[^[:space:]]' "$root/etc/trisse-home/"*.env; then echo "secret in config"; fail=1; fi
if ls "$root"/etc/ssh/ssh_host_*key* >/dev/null 2>&1; then echo "SSH host keys present"; fail=1; fi
if awk -F: '$2 !~ /^[!*]/ && $2 != "" { found=1; print "password hash for " $1 } END { exit !found }' "$root/etc/shadow"; then fail=1; fi
if [ -s "$root/etc/machine-id" ] && [ "$(cat "$root/etc/machine-id")" != uninitialized ]; then echo "machine-id set"; fail=1; fi
[ "$fail" -eq 0 ] || { echo "image contains secrets" >&2; exit 1; }
echo "ok"

log "zero free space (smaller download)"
# Deleted package caches leave non-zero blocks that compress badly.
dd if=/dev/zero of="$root/zero.fill" bs=4M status=none 2>/dev/null || true
rm -f "$root/zero.fill"
dd if=/dev/zero of="$root/boot/firmware/zero.fill" bs=1M status=none 2>/dev/null || true
rm -f "$root/boot/firmware/zero.fill"
sync

log "finish"
cleanup
loop=""
trap - EXIT
size=$(stat -c %s "$image")
xz -T0 -6 -c "$image" >"$out/$name.img.xz"
(cd "$out" && sha256sum "$name.img.xz" >"$name.img.xz.sha256")
"$here/os-list.sh" "$out/$name.img.xz" "$size" "$version" >"$out/os_list.json"
# Hand the results to the calling user (CI uploads them without sudo).
chown -R "${SUDO_UID:-0}:${SUDO_GID:-0}" "$out"
ls -l "$out"
echo "$image" >"$work/last-image"
