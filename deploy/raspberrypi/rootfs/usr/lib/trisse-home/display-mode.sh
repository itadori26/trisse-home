#!/bin/bash
# Applies the display mode requested by trisse-home (#140). Runs as root from
# trisse-home-display.service; the app itself stays unprivileged and can only
# write the request file. Accepts exactly "headless" or "kiosk".
set -uo pipefail

dir=${TRISSE_DEVICE_DIR:-/var/lib/trisse-home/device}
env_file=${TRISSE_KIOSK_ENV:-/etc/trisse-home/kiosk.env}
dry=${TRISSE_DISPLAY_DRY_RUN:-}

status() { # status <mode> <phase> [error]
	# The directory belongs to the unprivileged service: never write through
	# a name it could have prepared (symlink). mktemp creates a fresh file
	# (O_EXCL, random name) and mv replaces the target without following it.
	local err=${3:-} tmp
	tmp=$(mktemp "$dir/.display-mode.status.XXXXXX") || return
	printf '{"mode":"%s","phase":"%s","error":"%s"}\n' "$1" "$2" "${err//\"/\'}" >"$tmp"
	chmod 0644 "$tmp"
	mv -fT "$tmp" "$dir/display-mode.status"
}
# readfile <name> <max bytes>: a regular file of the service, never a link or FIFO.
readfile() {
	local f="$dir/$1"
	[ -f "$f" ] && [ ! -L "$f" ] || return 0
	head -c "$2" "$f"
}
run() { if [ -n "$dry" ]; then echo "would run: $*"; else "$@"; fi; }

[ -d "$dir" ] && [ ! -L "$dir" ] || exit 0
mode=$(readfile display-mode 16 | tr -d '[:space:]')
case "$mode" in
headless | kiosk) ;;
*)
	status "" failed "invalid request"
	exit 1
	;;
esac

status "$mode" installing
if [ "$mode" = kiosk ]; then
	url=$(readfile kiosk-url 512 | tr -d '[:space:]')
	if [[ ! "$url" =~ ^http://localhost/kiosk#[A-Za-z0-9_-]+$ ]]; then
		status kiosk failed "no kiosk link"
		exit 1
	fi
	if ! command -v cage >/dev/null || ! command -v chromium >/dev/null; then
		if ! run apt-get update -qq || ! run env DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends cage chromium; then
			status kiosk failed "could not install the display packages (internet connection?)"
			exit 1
		fi
	fi
	id trisse-kiosk >/dev/null 2>&1 || run useradd --system --create-home --shell /usr/sbin/nologin --groups video,render,input trisse-kiosk
	umask 027
	printf 'KIOSK_URL=%s\n' "$url" >"$env_file.tmp" && mv "$env_file.tmp" "$env_file"
	chgrp trisse-kiosk "$env_file" 2>/dev/null || true
	run systemctl set-default graphical.target
	if ! run systemctl enable --now trisse-home-kiosk.service; then
		status kiosk failed "the display could not be started"
		exit 1
	fi
	run systemctl restart trisse-home-kiosk.service
else
	run systemctl disable --now trisse-home-kiosk.service 2>/dev/null
	rm -f "$env_file"
	run systemctl set-default multi-user.target
fi
status "$mode" active
