#!/bin/sh
# Exports the WireGuard peer status for the trisse-home "wireguard" widget.
#
# Run it every minute on the Docker host, e.g. /etc/cron.d/trisse-home-wireguard:
#   * * * * * root /opt/stacks/trisse-home/scripts/wireguard-export.sh
#
# No secret is written: only peer lines of `wg show all dump` are kept and the
# preshared key column is blanked; the interface line (private key) is dropped.
# Peer names are read from the public keys in the peer_<name> folders of
# linuxserver/wireguard (optional).
#
# Environment:
#   WG_CONTAINER  container running WireGuard (default: wireguard); set
#                 WG_CONTAINER= (empty) to call wg on the host itself
#   OUT_DIR       target directory, mounted into trisse-home as /wireguard
#                 (default: the "wireguard" directory next to this script's parent)
set -eu

CONTAINER=${WG_CONTAINER-wireguard}
OUT_DIR=${OUT_DIR:-$(dirname "$0")/../wireguard}
umask 022
mkdir -p "$OUT_DIR"

run() {
	if [ -n "$CONTAINER" ]; then
		docker exec "$CONTAINER" "$@"
	else
		"$@"
	fi
}

tmp=$(mktemp "$OUT_DIR/.export.XXXXXX")
trap 'rm -f "$tmp" "$tmp.names"' EXIT

# Peer lines have 9 tab-separated fields; field 3 is the preshared key.
run wg show all dump | awk -F '\t' 'BEGIN { OFS = "\t" } NF == 9 { $3 = "-"; print }' >"$tmp"

# "<public key> <name>" for linuxserver/wireguard peer folders (best effort).
run sh -c 'for d in /config/peer_*; do
	[ -d "$d" ] || continue
	n=${d##*/peer_}
	f="$d/publickey-peer_$n"
	[ -f "$f" ] && printf "%s %s\n" "$(cat "$f")" "$n"
done' >"$tmp.names" 2>/dev/null || true

chmod 644 "$tmp" "$tmp.names"
mv "$tmp.names" "$OUT_DIR/peer-names.txt"
mv "$tmp" "$OUT_DIR/peers.dump"
trap - EXIT
