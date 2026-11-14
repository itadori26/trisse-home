# Device (display mode)

**What it does:** On the Raspberry Pi appliance image (ADR 0015), you choose whether the Pi is **only a server** or also **shows the dashboard full-screen** on a connected display: HDMI or the official touch display (#140). Docker and plain `.deb` installs don't show this at all.

## Choosing the mode

After the first login the appliance asks once: *How do you want to use this device?* You can change the answer any time under **Settings → Device**. The step says whether a display is connected (from `/sys/class/drm/*/status`).

**Server + display** does the following:
- installs `cage` (a minimal Wayland kiosk compositor) and Chromium from the Raspberry Pi OS repository, so the device needs internet for that;
- creates an unprivileged `trisse-kiosk` user with auto-login;
- starts Chromium in kiosk mode on `http://localhost/kiosk#…`.

The display signs in with a kiosk link of its own (`this-device` under *Kiosk displays*). Its token is **accepted only from the device itself (loopback)**. Rotation and night dimming can be set there like for any kiosk display. **Server only** stops and disables the display and removes its kiosk link.

## How it stays least-privileged

The app runs unprivileged (`ProtectSystem=strict`, no capabilities) and stays that way. It can only write files into `/var/lib/trisse-home/device/`:

| File | Written by | Content |
| --- | --- | --- |
| `display-mode` | app | `headless` or `kiosk` |
| `kiosk-url` | app (mode 0600) | `http://localhost/kiosk#<token>` |
| `display-mode.status` | root script | `{"mode":"kiosk","phase":"installing\|active\|failed","error":"…"}` |

A root-owned `trisse-home-display.path` unit watches `display-mode` and runs `/usr/lib/trisse-home/display-mode.sh`. That fixed script accepts only `headless` or `kiosk`, does the installation and writes the status. The settings page shows the progress and errors, such as no internet for the download.

## Cost

Nothing on the default image. The display adds several hundred MB of storage and the RAM of a browser, but only when you choose it (ADR 0011). For a Pi Zero 2 W or a Pi 3, *server only* is recommended.

## Not yet verified

The root script is tested in CI only for its validation. **Installing the kiosk on real hardware (Pi 4/5 with HDMI and the touch display) and measuring RAM/storage is still to be done.** See #140.

## API

`GET /api/device`, `PUT /api/device/display-mode`. See `api/modules/device.yaml`.
