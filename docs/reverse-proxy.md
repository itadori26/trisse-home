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

## Single sign-on (optional)

### OpenID Connect (Authelia, Authentik, Keycloak, Pocket ID, …)

Register a confidential client with the redirect URI
`https://home.example.org/api/auth/oidc/callback`, then add to `.env`:

```sh
OIDC_ISSUER=https://auth.example.org
OIDC_CLIENT_ID=trisse-home
OIDC_CLIENT_SECRET=…
OIDC_REDIRECT_URL=https://home.example.org/api/auth/oidc/callback
OIDC_ALLOWED_EMAILS=me@example.org        # or OIDC_ALLOWED_SUBJECTS=<sub>
# PASSWORD_LOGIN=false                    # once SSO works
# SSO_REQUIRE_2FA=true                    # also ask for the local TOTP code
```

The login page then shows **Sign in with single sign-on**. Authelia example
(`configuration.yml`):

```yaml
identity_providers:
  oidc:
    clients:
      - client_id: trisse-home
        client_secret: '$pbkdf2-sha512$…'   # hashed secret
        authorization_policy: two_factor
        redirect_uris: [https://home.example.org/api/auth/oidc/callback]
        scopes: [openid, email, profile]
        token_endpoint_auth_method: client_secret_basic
```

In Authentik create an *OAuth2/OpenID Provider* (confidential, the same
redirect URI, signing key RS256) and an application; the issuer is
`https://auth.example.org/application/o/trisse-home/`.

### Forward auth with a trusted header

If the proxy already authenticates every request (Authelia/Authentik
forward auth), trisse-home can trust the user name it passes on:

```sh
TRUSTED_PROXIES=172.20.0.0/16     # the proxy network - required
AUTH_PROXY_HEADER=Remote-User
AUTH_PROXY_USERS=me
```

The header is **only** believed when the request comes directly from an
address in `TRUSTED_PROXIES`; the proxy must remove the header from client
requests (Authelia and Authentik do). Traefik example:

```yaml
http:
  middlewares:
    authelia:
      forwardAuth:
        address: http://authelia:9091/api/authz/forward-auth
        authResponseHeaders: [Remote-User, Remote-Email]
```

Recovery: if the identity provider is down and the password login is
disabled, run `docker compose exec trisse-home /trisse-home reset-account`
and set a new password with the setup token from the log.

## Passkeys

Passkeys (Settings → Account) need a secure context and a fixed origin. Set
the address users open:

```sh
PUBLIC_URL=https://home.example.org
```

Then add a passkey in Settings → Account; the next login asks for it after
the password. TOTP and recovery codes keep working as alternatives.

