# API tokens

Personal API tokens let scripts and automations use the trisse-home API
without a browser session: iOS Shortcuts, Home Assistant, cron jobs.

## Create a token

Settings → **API tokens** → choose a name, the access per module (*read* or
*read & write*) and optionally an expiry. The token (`th_…`) is shown **once**
— copy it into your automation. Only its hash is stored; it can be revoked
at any time, and its last use (time and IP) is shown in the list and logged
in the security log.

Tokens can never reach the account endpoints (`/api/auth/*`): a leaked token
cannot create more tokens, change the password or read the security log.
Requests with a token still have to come from `ALLOWED_CIDRS`, and invalid
tokens count towards the login lockout.

## Use a token

Send it in the `Authorization` header; no cookie and no CSRF token are
needed:

```sh
curl -H "Authorization: Bearer th_…" http://192.168.178.115:8090/api/todos

curl -X POST -H "Authorization: Bearer th_…" -H "Content-Type: application/json" \
  -d '{"title":"Milch kaufen"}' http://192.168.178.115:8090/api/todos
```

All endpoints are described in OpenAPI: [`api/openapi.base.yaml`](../api/openapi.base.yaml)
plus one fragment per module in [`api/modules/`](../api/modules/); `make openapi` merges
them into `dist/openapi.yaml`.

## Example: iOS Shortcut "Add to-do"

Create a token with **todos: read & write**, then build a shortcut:

1. **Ask for Input** — prompt "To-do", input type *Text*.
2. **Get Contents of URL**
   - URL: `http://<your-server>:8090/api/todos`
   - Method: `POST`
   - Headers: `Authorization` = `Bearer th_…`, `Content-Type` = `application/json`
   - Request Body: *JSON* with the key `title` set to *Provided Input*
3. Optional: **Show Notification** "Added".

Run it with Siri ("Add to-do"), from the share sheet or a home-screen icon.
The phone must be in your LAN or VPN, as for the web interface.

## Example: Home Assistant REST command

```yaml
rest_command:
  trisse_todo:
    url: "http://192.168.178.115:8090/api/todos"
    method: POST
    headers:
      Authorization: !secret trisse_home_token
      Content-Type: application/json
    payload: '{"title": "{{ title }}"}'
```

with `trisse_home_token: "Bearer th_…"` in `secrets.yaml`.
