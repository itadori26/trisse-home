# Native install (Debian, Ubuntu, Raspberry Pi OS)

trisse-home is a single static binary, so it does not need Docker. Every
release attaches `.deb` packages for `amd64` and `arm64` that install it as a
hardened systemd service. This is also what the Raspberry Pi image is built
from ([ADR 0015](adr/0015-raspberry-pi-image.md)).

Docker stays the primary setup on a server that already runs containers; the
native package is for machines without Docker and for small single-purpose
devices.

## Install

Pick the package for your architecture (`dpkg --print-architecture`) from the
[latest release](https://github.com/itadori26/trisse-home/releases/latest),
check it and install it:

```sh
VERSION=1.2.3   # the release you want
ARCH=$(dpkg --print-architecture)
curl -fLO "https://github.com/itadori26/trisse-home/releases/download/v$VERSION/trisse-home_${VERSION}_$ARCH.deb"
curl -fLO "https://github.com/itadori26/trisse-home/releases/download/v$VERSION/trisse-home_${VERSION}_packages.sha256"
sha256sum --check --ignore-missing "trisse-home_${VERSION}_packages.sha256"
sudo apt install "./trisse-home_${VERSION}_$ARCH.deb"
```

The service starts right away on port 8090. Open `http://<host>:8090` and
enter the setup token from the log:

```sh
sudo journalctl -u trisse-home | grep setup_token
```

## What the package installs

| Path | Purpose |
| --- | --- |
| `/usr/bin/trisse-home` | The binary (same as in the container image) |
| `/etc/trisse-home/trisse-home.env` | Configuration (environment variables, see [.env.example](../.env.example)) |
| `/etc/trisse-home/secret.env` | `SECRET_KEY`, generated on the first start |
| `/var/lib/trisse-home/` | Database and automatic backups |
| `trisse-home.service` | systemd unit, runs as the system user `trisse-home` |

**`SECRET_KEY` is created on the device** on the first start (32 random bytes)
— it is never part of a package or an image. It encrypts stored integration
passwords: keep a copy of `/etc/trisse-home/secret.env` with your backups.

After changing the configuration: `sudo systemctl restart trisse-home`.

Defaults differ from Docker where Docker needed workarounds: host metrics are
read from `/proc` and `/sys` directly, Wake-on-LAN broadcasts without a relay,
and the Docker module is disabled (`MODULES_DISABLED=docker`) because there is
no Docker socket proxy.

## Sandbox

The service runs without capabilities and with a read-only view of the
system; only `/var/lib/trisse-home` is writable. Check it with
`systemd-analyze security trisse-home` (CI keeps the exposure at or below
2.0). Changes go into a drop-in (`sudo systemctl edit trisse-home`), never
into the packaged unit:

```ini
# Listen on port 80 (also set LISTEN_ADDR=:80 in trisse-home.env)
[Service]
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE
```

```ini
# Backups on a USB stick (also set BACKUP_DIR=/mnt/usb/trisse-home in trisse-home.env)
[Service]
ReadWritePaths=/mnt/usb/trisse-home
```

## Robustness

- **Watchdog:** the service reports readiness to systemd and pings its
  watchdog every 30 seconds while the database answers. If it hangs, systemd
  restarts it after 60 seconds (`WatchdogSec=`).
- **Power loss:** SQLite runs in WAL mode with `synchronous=NORMAL`. After a
  sudden power cut the database is always consistent; at worst the last
  change before the cut is lost. Daily backups cover the rest.

## Update and rollback

Install the newer package the same way; configuration, key and data are kept
and the service restarts. To roll back, install the older package with
`sudo apt install --allow-downgrades ./trisse-home_<old>_<arch>.deb` —
database migrations only add columns and tables, and a backup is created
daily. (Automatic updates through an APT repository are planned in #137.)

## Forgot the password?

```sh
sudo -u trisse-home sh -c 'set -a; . /etc/trisse-home/secret.env; . /etc/trisse-home/trisse-home.env; exec trisse-home reset-account'
sudo systemctl restart trisse-home
sudo journalctl -u trisse-home | grep setup_token
```

## Moving from Docker

1. Stop the container and copy the database out of the volume (see *Backup*
   in the README) to `/var/lib/trisse-home/trisse-home.db`.
2. Put your existing `SECRET_KEY` into `/etc/trisse-home/trisse-home.env` — a
   key there always wins over the generated one.
3. `sudo chown -R trisse-home:trisse-home /var/lib/trisse-home` and
   `sudo systemctl restart trisse-home`.

## Uninstall

- `sudo apt remove trisse-home` stops the service and keeps configuration and
  data.
- `sudo apt purge trisse-home` also deletes `/etc/trisse-home`. The data in
  `/var/lib/trisse-home` is **never deleted automatically**; the generated key
  is moved next to it (`secret.env.purged`) so a reinstall can still decrypt
  it. Delete the directory yourself if you no longer need it.
