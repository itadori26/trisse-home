# Running behind a reverse proxy

By default trisse-home is reached directly via `http://<server-ip>:8090` and
protected by the IP allowlist. To serve it under a domain with HTTPS, put it
behind a reverse proxy. The example uses
[nginx-proxy](https://github.com/nginx-proxy/nginx-proxy) with
[acme-companion](https://github.com/nginx-proxy/acme-companion), but any
proxy that sets `X-Forwarded-For` works.

## 1. Join the proxy network

Copy [`compose.proxy.example.yaml`](../compose.proxy.example.yaml) to
`compose.override.yaml`, set your domain and the name of the network the
proxy container is attached to. Compose merges the override automatically.

## 2. Adjust `.env`

```dotenv
# HTTPS is terminated by the proxy: mark the session cookie Secure and use
# the __Host- prefix.
COOKIE_SECURE=true

# Requests now come from the proxy. Trust its X-Forwarded-For header so the
# allowlist and rate limiting see the real client IP. Use the subnet of the
# proxy network (docker network inspect <network>).
TRUSTED_PROXIES=172.30.0.0/24

# Keep the allowlist restricted to networks that may use the app.
ALLOWED_CIDRS=192.168.178.0/24,10.13.13.0/24
```

The proxy's own address must **not** be in `ALLOWED_CIDRS` unless every
client behind the proxy should be allowed.

## 3. Optionally stop publishing the port

Once the proxy works, remove the `ports:` mapping (or bind it to
`127.0.0.1`) so the app is only reachable through the proxy.

## Notes

- The `Host` header is forwarded by nginx-proxy; the OpenSearch description
  uses it together with `X-Forwarded-Proto` to build absolute URLs.
- Go's cross-origin protection compares `Origin` with the request host, so
  the proxy must not rewrite the `Host` header.
- Sessions created over plain HTTP are not valid after switching
  `COOKIE_SECURE` because the cookie name changes; simply log in again.
