#!/bin/sh
# postinst for the trisse-home .deb. The systemd handling follows what
# dh_installsystemd generates, so the package behaves like any Debian service:
# enabled and started on install, restarted on upgrade, an admin's
# "systemctl disable" is respected.
set -e

# The service and the timer that shows the setup token on a headless device.
UNITS="trisse-home.service trisse-home-setup-info.timer"

create_user() {
	if getent passwd trisse-home >/dev/null; then
		return
	fi
	if command -v systemd-sysusers >/dev/null 2>&1; then
		systemd-sysusers trisse-home.conf
	else
		useradd --system --user-group --home-dir /var/lib/trisse-home \
			--no-create-home --shell /usr/sbin/nologin \
			--comment "trisse-home start page" trisse-home
	fi
}

fix_permissions() {
	# The config may contain credentials (SECRET_KEY, OIDC secret): readable
	# by root and the service group only.
	chown root:trisse-home /etc/trisse-home /etc/trisse-home/trisse-home.env
	chmod 0750 /etc/trisse-home
	chmod 0640 /etc/trisse-home/trisse-home.env
	if [ -f /etc/trisse-home/secret.env ]; then
		chown root:trisse-home /etc/trisse-home/secret.env
		chmod 0640 /etc/trisse-home/secret.env
	fi
}

case "$1" in
configure | abort-upgrade | abort-deconfigure | abort-remove)
	create_user
	fix_permissions

	for unit in $UNITS; do
		# Only removes masks created by deb-systemd-helper on package removal.
		deb-systemd-helper unmask "$unit" >/dev/null || true
		# was-enabled defaults to true, so new installations get enabled.
		if deb-systemd-helper --quiet was-enabled "$unit"; then
			deb-systemd-helper enable "$unit" >/dev/null || true
		else
			deb-systemd-helper update-state "$unit" >/dev/null || true
		fi
	done

	# No running systemd (image build in a chroot, container): enabled for the
	# next boot only. deb-systemd-invoke also honours policy-rc.d.
	if [ -d /run/systemd/system ]; then
		systemctl --system daemon-reload >/dev/null || true
		if [ -n "$2" ]; then
			deb-systemd-invoke restart trisse-home.service >/dev/null || true
		else
			deb-systemd-invoke start $UNITS >/dev/null || true
			echo "trisse-home is starting. Show the address and setup token with:"
			echo "  sudo trisse-home-setup-token"
		fi
	fi
	;;
esac

exit 0
