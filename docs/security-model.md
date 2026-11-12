# Security model

trisse-home runs on a home server next to things that matter: the Docker
daemon, the WireGuard VPN, the router. The guiding rule is **least privilege
per component**: the application gets read access to exactly what a widget
shows, and every capability that could change something lives in a separate,
optional, minimal component that you enable on purpose.

This page explains why the optional setups are an extra step and what each
component can and cannot do.

## Trust boundaries

```mermaid
flowchart LR
  subgraph LAN["Your LAN / VPN (ALLOWED_CIDRS)"]
    browser["Browser\n(session cookie, CSRF token)"]
  end

  subgraph net["Docker network trisse-home"]
    app["trisse-home\nnon-root, read-only FS,\nno capabilities"]
    proxy["docker-socket-proxy\nGET only"]
    actions["docker-actions-proxy\n(optional) restart only"]
  end

  subgraph host["Docker host"]
    sock[("/var/run/docker.sock")]
    procfs["/proc, /sys, / (read-only)"]
    wgdir["./wireguard\n(peer status, no keys)"]
    cron["cron: wireguard-export.sh"]
    smartdir["./smart\n(smartctl JSON)"]
    smartcron["cron: smart-export.sh"]
    relay["wol-relay (optional)\nhost network, magic packets only"]
  end

  internet(("Internet APIs\nyou configured"))

  browser -- "HTTP(S) :8090" --> app
  app -- "GET /containers" --> proxy --> sock
  app -. "POST restart (allow-list)" .-> actions --> sock
  app -- "read" --> procfs
  app -- "read" --> wgdir
  cron -- "wg show dump\n(keys stripped)" --> wgdir
  app -- "read" --> smartdir
  smartcron -- "smartctl --json -a" --> smartdir
  app -. "magic packet" .-> relay -. "broadcast" .-> LAN
  app -- "only configured\nintegrations" --> internet
```

Solid lines are always present, dotted lines only when you enable the
optional component.

| Component | Can | Cannot |
| --- | --- | --- |
| **trisse-home** container | Read container list and stats through the read-only proxy, read host metrics from read-only mounts, read the WireGuard peer and S.M.A.R.T. exports, call the integrations you configured | Write to Docker, read files outside its mounts, gain privileges (non-root, no capabilities, `no-new-privileges`, read-only root FS), be reached from outside `ALLOWED_CIDRS` |
| **docker-socket-proxy** | Forward `GET` requests to the container and image endpoints (plus ping/version); `IMAGES=1` lets trisse-home read the repo digests and labels of the images of running containers for the update hints | Write anything (`POST: 0`) — it cannot pull, delete or build images — read system info, volumes, networks or secrets |
| **docker-actions-proxy** (profile `docker-actions`) | Forward `POST /containers/<name>/stop\|restart\|kill` | Any read, create, exec, image or volume operation; trisse-home itself only calls *restart* for containers on `DOCKER_ACTIONS_CONTAINERS` |
| **wireguard-export.sh** (cron, root) | Run `wg show all dump` and write peer public keys, endpoints, handshakes and traffic | Write private or preshared keys — the interface line is dropped and the PSK column blanked |
| **smart-export.sh** (cron, root) | Run `smartctl --scan-open` and `smartctl --json -a` and write one JSON file per disk | Change disks or read their data — smartctl only reads the health log; the container gets no device access |
| **wol-relay** (profile `wol`) | Accept valid magic packets on UDP 9009 and broadcast them into the LAN, rate-limited | Read data, Docker or secrets; forward anything that is not a magic packet; it has no capabilities and a read-only filesystem |
| **Stored credentials** | Be used by the module they belong to | Be read back through the API or appear in logs: they are encrypted with AES-256-GCM using `SECRET_KEY` and only returned as "set / not set" |

## Why each optional setup is an extra step

### Docker socket proxy

Whoever can talk to the Docker socket is root on the host: they can start a
privileged container that mounts `/`. Mounting the socket into a web
application would turn every bug in that application into a full host
compromise. The proxy reduces the API to the read endpoints the widget needs.

### Docker actions

Restarting needs `POST`, which the read-only proxy rightly refuses. Instead
of loosening it, a **second** proxy allows nothing but stop/restart/kill and
nothing else — not even reads — and trisse-home adds its own allow-list of
container names on top. Logs are read through the read-only proxy.

### WireGuard export

`wg` needs `CAP_NET_ADMIN` in the network namespace of the VPN, and its
output contains the server's private key. Giving the dashboard either would
put the whole VPN at risk. A tiny script on the host, run by cron, writes only
the public peer status into a directory that is mounted read-only into
trisse-home. The script is short enough to read in a minute:
[`scripts/wireguard-export.sh`](../scripts/wireguard-export.sh).

### Prometheus metrics

`GET /metrics` exists only with `METRICS_ENABLED=true`. It needs a bearer
token of at least 32 characters (compared in constant time) and is limited
by `ALLOWED_CIDRS` like every other route. It reveals service, container and
disk names, but no credentials and nothing from your data. See
[metrics.md](metrics.md).

### S.M.A.R.T. export

Reading disk health needs raw device access (`/dev/sd*`, `CAP_SYS_RAWIO`),
which would let a compromised dashboard read every disk. Instead a host
script run by cron writes the JSON output of `smartctl` into a directory that
is mounted read-only into trisse-home:
[`scripts/smart-export.sh`](../scripts/smart-export.sh). The export contains
model and serial numbers and health counters, nothing from the disks' contents.

### Wake-on-LAN relay

Magic packets are broadcasts, and broadcasts do not leave a Docker bridge
network. Putting trisse-home itself into the host network would remove the
network isolation of the whole application. The relay is the only component
in the host network, and it can do exactly one thing: send magic packets —
which any device in your LAN can do anyway.

## FAQ: alternatives and their trade-offs

**Can I run trisse-home in the host network instead of using the relay?**
Yes: set `WOL_RELAY=` (empty). trisse-home then broadcasts itself. The
trade-off: the application can reach every service bound to `localhost` on
the host and loses the network isolation of its own Docker network.

**Can I avoid the extra container for Wake-on-LAN altogether?**
Enable directed broadcast forwarding on the host
(`net.ipv4.conf.all.bc_forwarding=1` and the same for your LAN interface),
set `WOL_RELAY=` and `WOL_BROADCAST=<lan>.255:9`. The trade-off: a host-wide
kernel setting that also affects every other container.

**Why not mount the Docker socket read-only (`:ro`)?**
`:ro` only prevents replacing the socket file; the API behind it is fully
writable. It gives no protection.

**Why not give the container `CAP_NET_ADMIN` for WireGuard?**
The capability is not limited to WireGuard: it allows changing routes,
firewall rules and interfaces of that network namespace. The export script
needs no change to the container at all.

**Can I expose trisse-home to the internet?**
It is designed for the LAN or a VPN. If you must, put it behind a reverse
proxy with HTTPS, set `COOKIE_SECURE=true`, `TRUSTED_PROXIES` and a strict
`ALLOWED_CIDRS`, and enable two-factor authentication. Never forward the port
in your router directly. See [reverse-proxy.md](reverse-proxy.md).

**What leaves my network?**
Only requests to services you configured (weather location, feeds, CalDAV,
DHL, …), each listed in the README section *Outbound requests*. There is no
telemetry; the update check is off by default.

## Reporting vulnerabilities

See [SECURITY.md](../SECURITY.md).
