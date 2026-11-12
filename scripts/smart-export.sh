#!/bin/sh
# Exports S.M.A.R.T. data for the trisse-home "diskhealth" widget.
#
# Run it on the Docker host as root (smartctl needs device access; the
# trisse-home container never gets any), e.g. /etc/cron.d/trisse-home-smart:
#   */30 * * * * root /opt/stacks/trisse-home/scripts/smart-export.sh
#
# Requires smartmontools 7.0 or newer (JSON output). For every disk found by
# `smartctl --scan-open` the output of `smartctl --json -a` is written to
# "<disk>.json"; files of disks that disappeared are removed. Only JSON is
# written, nothing is executed from the export directory.
#
# Environment:
#   OUT_DIR   target directory, mounted into trisse-home as /smart
#             (default: the "smart" directory next to this script's parent)
#   SMARTCTL  smartctl binary (default: smartctl)
set -eu

OUT_DIR=${OUT_DIR:-$(dirname "$0")/../smart}
SMARTCTL=${SMARTCTL:-smartctl}
umask 022
mkdir -p "$OUT_DIR"

# Lines look like: "/dev/sda -d sat # /dev/sda [SAT], ATA device".
"$SMARTCTL" --scan-open | while read -r dev opt type _; do
	case "$dev" in /dev/*) ;; *) continue ;; esac
	[ "$opt" = "-d" ] || type=auto
	name=$(printf '%s' "${dev#/dev/}" | tr -c 'A-Za-z0-9_.-' '_')
	case "$type" in *,*) name="${name}_$(printf '%s' "$type" | tr -c 'A-Za-z0-9_.-' '_')" ;; esac
	tmp=$(mktemp "$OUT_DIR/.export.XXXXXX")
	# smartctl's exit status is a bit mask that is non-zero for many
	# warnings; the JSON is evaluated by trisse-home instead.
	"$SMARTCTL" --json -a -d "$type" "$dev" >"$tmp" 2>/dev/null || true
	if [ -s "$tmp" ]; then
		chmod 644 "$tmp"
		mv "$tmp" "$OUT_DIR/$name.json"
		printf '%s\n' "$name.json"
	else
		rm -f "$tmp"
	fi
done >"$OUT_DIR/.written"

# Remove exports of disks that are gone.
for f in "$OUT_DIR"/*.json; do
	[ -e "$f" ] || continue
	grep -qxF "$(basename "$f")" "$OUT_DIR/.written" || rm -f "$f"
done
rm -f "$OUT_DIR/.written"
