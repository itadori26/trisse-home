# Read later

**What it does:** A read-later list (#99). Save a web page and trisse-home stores a cleaned **reader copy** (article text, headings, lists, quotes, code, tables, links and the article's images). The copy stays readable when the page changes or disappears. Articles are *unread*, *read* or *archived*, with an estimated reading time (220 words per minute). The widget lists unread articles with check-off.

## Saving

- The **+ Read later** field of the widget or Settings → Read later.
- The **quick capture page** (`/capture`, bookmarklet, the share target of the installed app and the browser extension) offers *Read later* as a target.
- The **RSS widget** has a "read later" button per item.

## Security

- The page is parsed with a real HTML5 parser (`golang.org/x/net/html`). The copy is **rebuilt from an allowlist** of elements: no scripts, styles, event handlers, forms, frames, SVG or MathML. Links are http(s) only and open with `rel="noopener noreferrer nofollow"`.
- Images are downloaded **once** by the server (raster formats only, no SVG). They are served from trisse-home with `nosniff` and a sandboxing CSP, so the browser never contacts the original site when you read.
- Page and image requests use the guarded client: no loopback or link-local/metadata addresses, checked on every connection and redirect. Addresses are validated centrally (#191).

## Limits

The reader copy is at most 1 MB of HTML. Per article at most 20 images, each at most 2 MB and 8 MB in total. All copies together are at most 200 MB; beyond that, saving is refused until archived articles are deleted. Data exports contain the articles and copies but not the images (backups contain everything).

## Outbound requests

Only the page you save and its images, once, when you save it.

## API

`GET/POST /api/readlater`, `GET/DELETE /api/readlater/{id}`, `PUT /api/readlater/{id}/status`, `GET /api/readlater/{id}/images/{n}` – see `api/modules/readlater.yaml`.
