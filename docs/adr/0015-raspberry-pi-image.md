# ADR 0015: Raspberry Pi appliance image

- Status: Proposed (becomes Accepted once the image was flashed with
  Raspberry Pi Imager on real hardware, see decision 7)
- Date: 2026-11-12
- Tracking: #130

## Context

trisse-home runs wherever Docker and Compose run, and multi-arch images
(including `arm64`) are published and smoke-tested for every release (#43).
That still assumes a Linux server, a terminal and a hand-edited `.env`.

Many people who would like a start page for their home own a spare
Raspberry Pi but no server. For them the shortest path is an *appliance*:
flash an SD card, boot, open a URL. The app is a good fit for that — one
static binary without CGO, SQLite in a single file, about 75 MB of memory
when idle (ADR 0011) — so the work is almost entirely about packaging and
operations, not about the application.

Constraints on a Raspberry Pi differ from a home server:

- **Storage** is usually an SD card: limited write endurance, corruption on
  power loss, slow random writes.
- **Memory** can be as low as 512 MB (Pi Zero 2 W).
- **Headless** is the norm: no keyboard or monitor during setup.
- **Public image:** every device is flashed from the same file, so the image
  must not contain any secret.
- **Updates** must work without reflashing, or devices will run old
  versions forever.

## Decision

1. **Base OS: Raspberry Pi OS Lite (64-bit), currently Debian Trixie.**
   It supports every 64-bit board, gets kernel and firmware updates from
   Raspberry Pi, and includes cloud-init, which Raspberry Pi Imager 2.x uses
   for its customisation (user, Wi-Fi, SSH key, locale, time zone). Users
   get the familiar Imager flow instead of a project-specific one.
   Supported boards: Pi 3B/3B+, 4, 400, 5, 500 and Zero 2 W.

2. **Runtime: native systemd service, no Docker in the image.**
   The binary is static, so a container adds nothing but weight on a
   single-purpose device: a second runtime in memory, image layers on the
   SD card and a second update path. Isolation comes from systemd instead:
   a dedicated system user, `ProtectSystem=strict`, `PrivateDevices`, a
   system call filter and an empty capability set (only
   `CAP_NET_BIND_SERVICE` to listen on port 80). Without Docker the
   workarounds needed for the container disappear: Wake-on-LAN broadcasts
   directly (no relay), the system module reads `/proc` and `/sys`
   directly. The Docker module is disabled by default on the image.

3. **Packaging: `.deb` packages, built with nfpm, for `arm64` and `amd64`.**
   The image installs the same package that any Debian or Ubuntu user can
   install without Docker. nfpm is a build-time tool only (ADR 0011).
   Configuration stays environment-based (`/etc/trisse-home/trisse-home.env`
   loaded with `EnvironmentFile=`), so `internal/platform/config` does not
   change. Data lives in `/var/lib/trisse-home`.

4. **Updates: signed APT repository plus `unattended-upgrades`.**
   Packages are published to an APT repository on GitHub Pages, signed with
   a dedicated key held in GitHub secrets. The image trusts that key via
   `Signed-By` and installs OS security updates and trisse-home releases
   automatically, with reboots only inside a maintenance window. Pinning a
   version and rolling back work with `apt`, mirroring
   `TRISSE_HOME_VERSION` for Docker.

5. **First boot creates all secrets on the device.**
   A oneshot service generates `SECRET_KEY` and a setup token before the
   app starts for the first time. The setup token is offered only through
   channels that need physical access or the owner's credentials: the HDMI
   console, a file on the boot partition of the SD card, `sudo` over SSH
   and the journal. Both visible copies are removed once the account
   exists. CI verifies that the image contains no `SECRET_KEY`, SSH host
   keys or password hashes.

6. **Built and tested in CI on `ubuntu-24.04-arm`.**
   Native arm64 runners are free for public repositories, so the image is
   built without emulation in the release workflow, boot-tested, and
   attached to the release with checksum, SBOM and provenance attestation.
   A Raspberry Pi Imager repository file (`os_list.json`,
   `init_format: cloudinit-rpi`) lets users pick the image inside Imager.

7. **Build method: customise the official Raspberry Pi OS Lite image**
   (result of the spike, #132). `deploy/raspberrypi/build-image.sh` starts
   from the released Lite image, pinned by URL and SHA-256 in `base.env`,
   grows the root partition, installs the package in a chroot on a native
   arm64 runner (no emulation), adds the files from `rootfs/` and zeroes
   the free space. The result is byte for byte Raspberry Pi OS Lite plus a
   small, reviewable layer, so cloud-init, the first-boot resize and Imager's
   customisation behave exactly as on the official image.

   Measured in CI on `ubuntu-24.04-arm` (2026-11-12): image build 220–260 s,
   boot test 25 s, whole job about 6.5 minutes; `.img.xz` 557.5 MB
   (the base is 525 MB). The size is a budget in `budgets.json`.

   Not chosen: **pi-gen** rebuilds Raspberry Pi OS from packages, so every
   build is much longer and the result only approximates the released image;
   **rpi-image-gen** starts from its own minimal Debian base with a
   different first-boot path and is still under active development. Neither
   was built in the spike: the customised image already met every goal, and
   switching later only touches `deploy/raspberrypi/`.

   Still to verify on hardware before this ADR is accepted: flashing with
   Raspberry Pi Imager 2.x including its customisation (Wi-Fi, user, SSH,
   hostname) on a Pi 4 or 5.

8. **SD-card care.** journald size-limited, `noatime`, swap as Raspberry
   Pi OS configures it, the existing WAL + `synchronous=NORMAL` SQLite
   settings (consistent after a power loss, at worst the last transaction
   is lost), and a systemd watchdog through a small `sd_notify` helper in
   `internal/platform` that is a no-op outside systemd (#136).

## Alternatives considered

- **Docker inside the image (Raspberry Pi OS + Docker + `compose.yaml`).**
  Least new code and identical to the documented setup, but `dockerd` and
  `containerd` permanently in memory, more SD-card writes, two update paths (apt and images) and the
  WoL relay and `/host` mounts that only exist because of the container.
  Users who want Docker can keep using the Compose setup on any Pi.
- **DietPi or Armbian as base.** Good optimisations for small boards, but
  their own first-boot systems and update tools; Imager customisation would
  not work the same way, and a third party would sit between us and kernel
  updates.
- **Buildroot or Yocto (read-only, A/B updates).** The most robust
  appliance, but a large maintenance burden for a two-person project, no
  `apt`, and every package has to be built from source.
- **Only document "install the .deb on Raspberry Pi OS".** Useful and part
  of this decision (decision 3), but still requires a terminal; it does not
  reach the people the image is for.

## Consequences

- A second distribution channel to maintain: packaging, APT repository,
  image build and the signing key. The release workflow grows by two jobs.
- The application code stays unchanged except for the generic systemd
  notify helper; the dependency rule (ADR 0001) is unaffected because all
  Pi-specific files live in `packaging/` and `deploy/raspberrypi/`.
- Native `.deb` packages also benefit users on other Debian-based machines
  who do not want Docker.
- The image size becomes a budget in `budgets.json` (ADR 0011).
- The Docker setup remains the primary, fully supported way to run
  trisse-home on a server; documentation presents both.
- **One image, two display modes (#140).** After the account setup, Pi
  installs ask whether the device is a server only (default) or should also
  show the dashboard full-screen on an attached display; the choice can be
  changed later in the settings. Kiosk components (a minimal Wayland
  compositor and browser) are a separate package installed on demand, so
  the default image stays lean. The app stays unprivileged: it can only
  write a requested mode into its state directory, and a root-owned systemd
  path unit applies it with a fixed, validating script.
