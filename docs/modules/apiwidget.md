# JSON widget

**What it does:** Shows values from any JSON API: the widget polls a configured endpoint, picks values out of the response with JSONPath and shows them with label, unit and optional warning/critical thresholds (#48). Useful for small home-lab services that have a JSON status page but no module of their own.

## Configuration

Settings → JSON APIs: per endpoint a name, the URL (http/https, query allowed), an optional request header such as an API key, the interval (30 seconds to 24 hours) and up to 10 values (label, JSONPath, unit, thresholds). *Test* shows the beginning of the response and the extracted values before saving. The widget then shows one endpoint.

Supported JSONPath: `$.a.b`, `$['name with spaces']`, `$.list[0]`, `$.list[-1]` and `$.list.length()`; the leading `$` may be left out. Numbers in strings count as numbers. A threshold is reached at or above its value, or at or below with *lower is worse* (e.g. free space).

## Security

- **Private addresses only by default:** loopback, RFC 1918, `fd00::/8` and `100.64.0.0/10`. Public addresses need *Allow public addresses* per endpoint. Link-local and cloud metadata addresses (`169.254.169.254`) are never allowed. The rule is checked on the resolved address of every connection, so DNS cannot get around it; redirects are not followed.
- The header value is stored encrypted with `SECRET_KEY`, never returned by the API and not included in data exports (the header then has to be entered again after an import).
- Responses are limited to 1 MB and 10 seconds; the settings preview shows at most 4 KB.

## Outbound requests

Only the configured endpoints, while a widget shows them, at most once per interval (failed requests are retried after a minute at the earliest). Nothing is requested until an endpoint is configured.

## API

`GET/POST /api/apiwidget/endpoints`, `PUT/DELETE /api/apiwidget/endpoints/{id}`, `GET /api/apiwidget/endpoints/{id}/values`, `POST /api/apiwidget/preview` – see `api/modules/apiwidget.yaml`. Guests do not see the widget's values yet.
