#!/bin/sh
# Pins the latest Raspberry Pi OS Lite (64-bit) in base.env, using the image
# list that Raspberry Pi Imager reads. Review the diff before committing.
#   deploy/raspberrypi/update-base.sh
set -eu

here=$(cd "$(dirname "$0")" && pwd)
list_url=https://downloads.raspberrypi.com/os_list_imagingutility_v4.json

curl -fsSL "$list_url" | jq -r '
	[.. | objects | select(has("url") and (.url | test("/raspios_lite_arm64/images/")))]
	| first
	| "BASE_NAME=\"\(.name), \(.release_date)\"\nBASE_URL=\"\(.url)\"\nBASE_SHA256=\"\(.image_download_sha256)\""
' >"$here/base.env.new"

{
	echo "# Base image of the Raspberry Pi appliance image (ADR 0015): the official"
	echo "# Raspberry Pi OS Lite (64-bit). Pinned by URL and SHA-256 so that every build"
	echo "# starts from exactly the same, verified file. Update with update-base.sh."
	cat "$here/base.env.new"
} >"$here/base.env"
rm -f "$here/base.env.new"
cat "$here/base.env"
