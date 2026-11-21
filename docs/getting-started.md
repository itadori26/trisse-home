# Getting started

This guide gets Waylight running and takes you through the first login.
It takes about ten minutes.

## Pick an installation

| You have … | Use | Guide |
| --- | --- | --- |
| A server that already runs Docker | Docker Compose | below |
| A Debian, Ubuntu or Raspberry Pi OS machine without Docker | The `.deb` package, a hardened systemd service | [Native install](native-install.md) |
| A spare Raspberry Pi and nothing else | The ready-made SD card image, updates itself | [Raspberry Pi](raspberry-pi.md) |

All three run the same program and store the same database, so you can move
between them later. Docker Compose is the most common setup and the one this
page describes.

## Install with Docker Compose

You need Docker with the Compose plugin. The repository is only used for
`compose.yaml`, the example configuration and a few optional helper scripts;
the image itself comes from `ghcr.io/waylight-app/waylight` (amd64 and
arm64). The previous name `ghcr.io/itadori26/waylight` stays available as an
alias with the same tags.

```sh
git clone https://github.com/Waylight-App/waylight.git
cd waylight
cp .env.example .env
sed -i "s|^SECRET_KEY=.*|SECRET_KEY=$(openssl rand -base64 32)|" .env
sudo docker compose up -d
```

`SECRET_KEY` encrypts the passwords and tokens you later enter for your
integrations. Keep a copy somewhere safe: a database backup is of limited use
without it.

Two settings in `.env` are worth a look before you go further:

- `BIND_ADDRESS`: set it to your server's LAN address so the port is not
  published on every interface.
- `ALLOWED_CIDRS`: the networks that may open the app. The default allows
  loopback and all private ranges. Narrow it if your LAN is shared.

Every other variable is described in the [configuration reference](configuration.md).

## Create your account

The first start prints a one-time setup token to the log:

```sh
sudo docker compose logs waylight | grep setup_token
```

Open `http://<server-ip>:8090`, enter the token and choose a password of at
least 12 characters. The token stops working as soon as the account exists.

Right after the login a few questions set up the essentials, one per
screen. Each shows its step ("Step 2 of 4"), and you can skip any of them or
close the dialog to skip the rest; everything can be changed later in the
settings.

1. **Your name** for the greeting on the start page.
2. **Your place**, found by search or with *Use my location* (the browser
   asks first). It is saved as the place *Home* and used by weather, rain
   radar, fuel prices and departures.
3. **Language.** Times follow the time zone of the device you use;
   Waylight has no time zone setting of its own.
4. **Look**: light, dark or like the device, and an accent colour. The page
   shows your choice right away.
5. **Name of this Waylight**, only when the branding of the organisation
   edition is licensed.

People you invite later get the personal questions only: name, language
and look.

Then Waylight asks what you use it for and offers four dashboard templates: Family, Homelab, Work and Wall tablet (see
[Templates](dashboard.md#templates)). Each one creates a tab laid out in
zones: *Today* on top, *My things* below and *System* at the bottom, which
only admins see. Widgets that have a compact view start small or as a
one-line strip, and the services strip feeds the attention line above the
zones, so a stopped service shows up there first. Templates with a weather
widget use the place from the questions above, or ask for one; you can also
set it later.

The first start also creates example data: to-dos, a welcome note,
bookmarks, a news feed and a few service tiles, so the widgets of a template
are not empty. Behind the templates sits an example dashboard with a *Home*
and a *Server* tab that shows every widget. Picking a template replaces it.
**Keep the example dashboard** keeps it instead, and **Start with an empty
dashboard** removes it and leaves a single tab with a clock. Closing the
dialog keeps the example dashboard as well. The example data of the modules
stays in every case; delete it whenever you like.

If you set `SAMPLE_DATA=false` before the first start, there is no example
data or example dashboard, and the same template choice starts from a single
tab with a clock.

Widgets and settings for a home server or network (Docker, Proxmox,
Pi-hole, …) are shown on a new installation. If you do not need them, set
**Settings → Appearance → The start page shows** to *Everyday*. Switching
only shows or hides widgets and settings and never deletes anything.

A short guided tour follows the template choice. You can restart it later
under Settings → About.

## Secure the account

Two steps are worth doing right away:

1. **Turn on two-factor authentication** under Settings → Account. You get
   ten recovery codes; store them outside Waylight.
2. **Do not forward the port on your router.** Use a VPN such as WireGuard to
   reach the dashboard from outside. If you want HTTPS, a public name or
   passkeys, put Waylight behind a [reverse proxy](reverse-proxy.md).

## Next steps

- [First steps](first-steps.md) shows you how to arrange tabs and widgets.
- [Backups](backups.md) explains what is backed up automatically and how to
  copy backups off the server.
- [Updates](updates.md) covers updating, pinning a version and rolling back.

## If something goes wrong

**The page does not load.** Check that the container is healthy with
`sudo docker compose ps`. If your browser runs on a network outside
`ALLOWED_CIDRS`, the server refuses the connection and logs the address it
rejected: `sudo docker compose logs waylight | grep -i denied`.

**You lost the setup token.** Run the `grep` command above again; the token
is printed on every start until the account exists.

**You forgot your password.** Reset the account and set it up again with a
new token. This removes the owner's sign-in data (password, two-factor,
passkeys, sessions); your tabs, widgets and everything in them stay.

```sh
sudo docker compose exec waylight /waylight reset-account
sudo docker compose restart waylight
sudo docker compose logs waylight | grep setup_token
```
