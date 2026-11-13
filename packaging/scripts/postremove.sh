#!/bin/sh
# postrm for the trisse-home .deb.
#
# remove: the unit is masked (dh_installsystemd behaviour), configuration and
#         data stay.
# purge:  the configuration in /etc/trisse-home is deleted. The data in
#         /var/lib/trisse-home is never deleted automatically; if it still
#         holds a database, the generated SECRET_KEY is moved next to it, so
#         a later reinstall can still decrypt stored credentials.
set -e

UNITS="trisse-home.service trisse-home-setup-info.timer"
STATE_DIR=/var/lib/trisse-home

case "$1" in
remove)
	if [ -x /usr/bin/deb-systemd-helper ]; then
		for unit in $UNITS; do
			deb-systemd-helper mask "$unit" >/dev/null || true
		done
	fi
	;;
purge)
	if [ -x /usr/bin/deb-systemd-helper ]; then
		for unit in $UNITS; do
			deb-systemd-helper purge "$unit" >/dev/null || true
			deb-systemd-helper unmask "$unit" >/dev/null || true
		done
	fi
	if [ -f /etc/trisse-home/secret.env ] && [ -f "$STATE_DIR/trisse-home.db" ]; then
		(umask 077 && cp /etc/trisse-home/secret.env "$STATE_DIR/secret.env.purged")
	fi
	rm -rf /etc/trisse-home
	rm -f /etc/issue.d/trisse-home.issue /boot/firmware/trisse-home-setup.txt
	if [ -d "$STATE_DIR" ]; then
		echo "trisse-home: data kept in $STATE_DIR (delete it yourself if you no longer need it)"
	fi
	;;
esac

if [ -d /run/systemd/system ]; then
	systemctl --system daemon-reload >/dev/null || true
fi

exit 0
