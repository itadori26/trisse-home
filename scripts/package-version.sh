#!/bin/sh
# Prints the Debian package version for the current git checkout.
#
#   v1.2.3 (tagged commit)        -> 1.2.3
#   4 commits after v1.2.3        -> 1.2.3+4.gabc1234   (sorts after 1.2.3)
#   no release tag yet            -> 0.0.0~dev+gabc1234 (sorts before 0.0.0)
set -eu

described=$(git describe --tags --match 'v[0-9]*' 2>/dev/null || true)
case "$described" in
v[0-9]*)
	# v1.2.3-4-gabc1234 -> 1.2.3+4.gabc1234
	echo "${described#v}" | sed 's/-/+/; s/-/./g'
	;;
*)
	echo "0.0.0~dev+g$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
	;;
esac
