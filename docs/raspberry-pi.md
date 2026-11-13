# trisse-home on a Raspberry Pi

The Raspberry Pi image turns a Pi into a small appliance for your start page:
flash an SD card, boot, open `http://trisse-home.local`. No Docker, no
terminal needed. It is Raspberry Pi OS Lite (64-bit) with trisse-home
installed as a hardened service ([ADR 0015](adr/0015-raspberry-pi-image.md));
everything runs on the Pi, your data never leaves it.

## What you need

| | |
| --- | --- |
| Board | Raspberry Pi 5, 500, 4, 400, 3B/3B+ or Zero 2 W (64-bit). Pi 1, 2 and Zero W are not supported. |
| SD card | 8 GB or more, a reputable brand (A1/A2). |
| Power | The official power supply of your board — weak supplies cause most "random" problems. |
| Network | Ethernet, or Wi-Fi configured while flashing. |

## 1. Flash the SD card

Install [Raspberry Pi Imager](https://www.raspberrypi.com/software/) (version
2.0 or newer).

**Recommended — Imager knows the image and offers its customisation:**

1. *App Options → Content Repository → Use custom URL*, enter
   `https://github.com/itadori26/trisse-home/releases/latest/download/os_list.json`
2. Choose your device, then **trisse-home** as operating system, then the SD
   card.
3. In the customisation step set at least a user and password, and Wi-Fi if
   the Pi has no cable. Hostname, time zone and SSH are optional. Write.

**Alternative — "Use custom" with the downloaded file:** download
`trisse-home-<version>-raspberrypi-arm64.img.xz` from the
[latest release](https://github.com/itadori26/trisse-home/releases/latest),
check it against the `.sha256` file and write it with *Use custom*. Imager may
not offer its customisation for a local file, so use Ethernet and add a user
later on the console if you need one. trisse-home itself needs no user
account on the Pi.

## 2. First start and setup

Insert the card, connect the network and power. The first boot takes one to
three minutes (the file system is enlarged, keys are generated). Then open

- `http://trisse-home.local` — or `http://<your-hostname>.local` if you set a
  hostname in Imager,
- or the Pi's IP address, which your router lists (e.g. FRITZ!Box: *Home
  Network → Network*). Some Android phones and older Windows versions do not
  resolve `.local` names.

The setup page asks for the **setup token**. It is created on the Pi during
the first boot, is different on every device, and works only until your
account exists. You find it

- on a monitor attached to the Pi, above the login prompt,
- in **`trisse-home-setup.txt` on the SD card** — shut the Pi down, put the
  card into any computer and open the file on the small `bootfs` drive,
- with `sudo trisse-home-setup-token` over SSH.

Choose a password (at least 12 characters) — done. The setup copies of the
token are deleted automatically afterwards. Want to choose the token
yourself? See *Setup token* in the [native install guide](native-install.md).

## Updates

The Pi installs updates by itself every night: security updates of Raspberry
Pi OS, kernel and firmware, and new trisse-home releases. If an update needs
a reboot (e.g. a new kernel), it reboots at 04:00. Settings → About shows the
running version.

Before every trisse-home update the database is copied to
`/var/lib/trisse-home/backups/pre-upgrade/`.

- Hold a version: `sudo apt-mark hold trisse-home` (`unhold` to continue).
- Roll back: `sudo apt install --allow-downgrades trisse-home=<version>` (the
  last five releases are available).

## Backups

trisse-home creates a database backup every day and keeps seven; download
them under Settings → Backups. **Keep the key as well** — without it stored
integration passwords (CalDAV, AdGuard Home, …) cannot be decrypted:

```sh
sudo cat /etc/trisse-home/secret.env
```

For a copy outside the SD card, put the backups on a USB stick (see *Sandbox*
in the [native install guide](native-install.md)) or use Settings → Export.

**Restore on a new SD card:** flash the image, do *not* finish the setup, then
over SSH: `sudo systemctl stop trisse-home`, copy your backup to
`/var/lib/trisse-home/trisse-home.db` and your old `SECRET_KEY` into
`/etc/trisse-home/trisse-home.env`, run
`sudo chown trisse-home:trisse-home /var/lib/trisse-home/trisse-home.db` and
`sudo systemctl start trisse-home`.

## HTTPS and passkeys

On the local network plain `http://` works. For access from outside use a VPN
(WireGuard, Tailscale) rather than forwarding a port. HTTPS — needed for
passkeys and to install trisse-home as an app — works with a reverse proxy
([guide](reverse-proxy.md)); set `PUBLIC_URL` and `COOKIE_SECURE=true` in
`/etc/trisse-home/trisse-home.env`.

## Configuration

Everything you would set in Docker's `.env` goes into
`/etc/trisse-home/trisse-home.env`; values there override the image defaults
(port 80, see `/etc/trisse-home/appliance.env`). Apply with
`sudo systemctl restart trisse-home`.

## Moving between the Pi and Docker

Settings → Export creates a JSON file with all your data (never secrets);
Settings → Import reads it on the other installation. To move everything
including history, copy the database and `SECRET_KEY` as described in
*Moving from Docker* in the [native install guide](native-install.md).

## Troubleshooting

| Problem | What to do |
| --- | --- |
| `trisse-home.local` does not open | Use the IP address from your router. Wait three minutes after the first boot. Check that phone or laptop are in the same network (not a guest Wi-Fi). |
| No network over Wi-Fi | Wi-Fi is only set when flashing with Imager's customisation; flash again or use a cable. |
| The Pi does not boot (no green LED activity) | Flash the card again, try another card and the official power supply. |
| Forgot the password | `sudo -u trisse-home sh -c 'set -a; . /etc/trisse-home/secret.env; . /etc/trisse-home/appliance.env; . /etc/trisse-home/trisse-home.env; exec trisse-home reset-account'`, then `sudo systemctl restart trisse-home` and `sudo trisse-home-setup-token`. |
| Something does not work | `systemctl status trisse-home` and `journalctl -u trisse-home -n 100`. The journal is limited to 50 MB to spare the SD card. |
| SD card full | `df -h /`; old backups are deleted automatically (`BACKUP_RETENTION`). |
