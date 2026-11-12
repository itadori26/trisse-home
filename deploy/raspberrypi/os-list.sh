#!/bin/sh
# Prints a Raspberry Pi Imager repository file (os_list.json) for the image, so
# users can pick trisse-home in Imager (App Options -> Content Repository) and
# still get Imager's customisation (user, Wi-Fi, SSH) through cloud-init.
#
#   os-list.sh <image.img.xz> <uncompressed size in bytes> <version>
set -eu

xz_file=${1:?image missing}
extract_size=${2:?size missing}
version=${3:?version missing}
repo=${GITHUB_REPOSITORY:-itadori26/trisse-home}

name=$(basename "$xz_file")
download_size=$(stat -c %s "$xz_file")
download_sha=$(sha256sum "$xz_file" | cut -d' ' -f1)
extract_sha=$(xz -dc "$xz_file" | sha256sum | cut -d' ' -f1)

jq -n \
	--arg version "$version" \
	--arg url "https://github.com/$repo/releases/download/v$version/$name" \
	--arg date "$(date -u +%Y-%m-%d)" \
	--argjson extract_size "$extract_size" \
	--argjson download_size "$download_size" \
	--arg download_sha "$download_sha" \
	--arg extract_sha "$extract_sha" \
	--arg icon "https://raw.githubusercontent.com/$repo/main/web/public/icons/icon-192.png" \
	'{
		os_list: [{
			name: ("trisse-home " + $version),
			description: "Self-hosted start page on Raspberry Pi OS Lite (64-bit). Open http://trisse-home.local after the first boot.",
			icon: $icon,
			url: $url,
			release_date: $date,
			extract_size: $extract_size,
			extract_sha256: $extract_sha,
			image_download_size: $download_size,
			image_download_sha256: $download_sha,
			init_format: "cloudinit-rpi",
			devices: ["pi5-64bit", "pi4-64bit", "pi3-64bit"]
		}]
	}'
