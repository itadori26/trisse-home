# Outbound requests

Waylight sends no telemetry and contacts no server of its own. It only
talks to services you set up or switch on. This page lists every outside
connection, when it happens and how often, so you can check the claim and
plan your firewall.

## Services you configure

These requests go to addresses that only the signed-in owner enters in the
settings, and only once a service is set up. Their passwords and tokens are
stored encrypted.

| Service | When and how often |
| --- | --- |
| Service tiles | A health check of each tile every minute; the TLS certificate is read from the same connection. |
| CalDAV calendars and task lists | Every 15 minutes, and when you add an event or item. |
| Google Calendar, Microsoft 365 | Only after you connect an account with your own app registration: the sign-in and token renewal (`accounts.google.com`, `oauth2.googleapis.com`, `login.microsoftonline.com`) and the shown calendars (`www.googleapis.com`, `graph.microsoft.com`) every 15 minutes, plus one request per new event ([details](modules/calendar.md)). |
| Wikis (BookStack, Outline, Wiki.js) | Only the wikis you connect, only while you search, at most four seconds per wiki ([details](modules/wiki.md)). |
| ICS subscriptions (calendar, waste, school holidays) | At their refresh interval, hourly by default, with `If-None-Match` and `If-Modified-Since`; school holidays at most weekly. |
| RSS, Atom and JSON feeds | Every 30 minutes. |
| AdGuard Home | Every minute. |
| Pi-hole | `/api/auth` once per session, then statistics and the blocking state every minute. |
| Nextcloud | Every five minutes. While you sign in from the setup assistant, `/index.php/login/v2/poll` at most every two seconds for up to ten minutes. |
| FRITZ!Box | Every 15 seconds; the call list only when a calls widget exists. |
| Home Assistant | `/api/states` while a smart home widget is open, at most every ten seconds. After signing in from the setup assistant, `/auth/token` to renew the access token shortly before it expires (every 30 minutes while the widget is in use). |
| Proxmox VE | `/api2/json/cluster/resources` every 30 seconds. |
| Sonarr, Radarr, Prowlarr | While an arr widget is open, at most once a minute. |
| Jellyfin, Plex | While a media widget is open, at most every 15 seconds. |
| Immich | While an Immich widget is open, at most every five minutes. |
| Speedtest Tracker | Every ten minutes. |
| JSON widget endpoints | While a widget shows them, at most once per configured interval. |
| Connected accounts (`ee` builds) | Only for members who connect an account: the sign-in, token renewal shortly before expiry and revocation on disconnect at `login.microsoftonline.com`, `auth.atlassian.com`, `api.notion.com` or `slack.com` ([details](connected-accounts.md)). |
| Team wikis (`ee` builds) | Only while a person searches and only the services they connected: `api.atlassian.com`, `graph.microsoft.com`, `api.notion.com`, `slack.com`, at most four seconds each ([details](modules/wiki.md#team-wikis-with-each-persons-account)). |
| Slack command (`ee` builds) | Only with `SLACK_APP_TOKEN`: one WebSocket connection to Slack, renewed hourly, and `slack.com` to open it and to match connected accounts ([details](slack-command.md)). |
| Teams command box search (`ee` builds) | Only with `TEAMS_BOT_APP_ID`: Microsoft's signing keys at `login.botframework.com` (daily) and `graph.microsoft.com` to match connected accounts ([details](teams.md)). |
| Teams tab (`ee` builds) | None from the server. The browser inside Teams loads Microsoft's Teams library from `res.cdn.office.net` for the tab page only ([details](teams.md#the-tab)). |
| Ollama | Only when you ask for a summary. |
| WebDAV off-site backups | After each backup, and when you restore from it. |
| ntfy, Gotify, webhooks, SMTP, Slack (`hooks.slack.com`), Teams (the workflow's address, usually `*.logic.azure.com` or `*.powerplatform.com`) | When a notification is sent, or an announcement is posted to a Slack or Teams channel. None until you add the channel. |
| Football results (`api.openligadb.de`) | Only while a sports widget is shown, at most every 15 minutes per league ([details](modules/sports.md)). |
| CardDAV address book (birthdays) | When you save the address book in the settings, then every six hours. None until you connect one ([details](modules/birthdays.md)). |
| CardDAV address book (phone list) | When an admin saves the address book in the settings, then every hour. None until one is connected ([details](modules/phonelist.md)). |
| Electricity prices (`api.awattar.de` / `api.awattar.at`) | Only while an electricity price widget is shown, at most every 15 minutes per bidding zone ([details](modules/electricity.md)). |
| Connected to-do lists (CalDAV) | While a to-do widget shows the list, at most once a minute: the list's CTag first, then only the to-dos whose ETag changed. None until you connect a list ([details](modules/todos.md)). |
| Markets (`www.ecb.europa.eu`, `stooq.com`) | Only while a markets widget is shown, at most once an hour per symbol; for currency conversion in the search, the ECB file at most once a day (`CURRENCY_RATES=false` switches that off) ([details](modules/markets.md)). |
| Todoist (`api.todoist.com`) | While a to-do widget shows a connected Todoist project: the sync endpoint at most every 15 seconds (only changes after the first request), plus one request per change you make. Pauses when Todoist answers 429. None until you connect Todoist ([details](modules/todos.md)). |
| IMAP inboxes (mail widget) | Only after you add an account: one TLS connection per account while a mail widget is shown, at most every two minutes; headers only, read-only ([details](modules/mail.md)). |
| Microsoft To Do, Google Tasks, TickTick | Only after you sign in with your own app registration: the sign-in and token renewal at the provider (`login.microsoftonline.com`, `oauth2.googleapis.com`, `ticktick.com`), and the list (`graph.microsoft.com`, `tasks.googleapis.com`, `api.ticktick.com`) while a to-do widget shows it, at most every 30 seconds, plus one request per change ([details](modules/todos.md)). |
| IMAP mailbox (parcels) | One folder, read-only over TLS, at the interval you choose. Off by default ([details](parcels-mail.md)). |

## Public services

These are contacted only when a module that needs them is in use.

| Service | When and how often |
| --- | --- |
| [Open-Meteo](https://open-meteo.com) | Weather for the places of your weather widgets. |
| Deutscher Wetterdienst | Weather warnings, rain radar, pollen and UV index for places in Germany. The radar image is fetched by the server, so your browser never contacts the DWD. Pollen and UV only when a widget shows them, at most hourly. |
| [Transitous](https://transitous.org) | Departures, cached for one minute per stop. The stop search while you type a name in the widget settings, and the stops near a saved place once when you pick that place there (the place's rounded coordinates are sent). Please respect its fair-use policy, or run your own MOTIS and set `TRANSIT_API_URL`. |
| DHL tracking API | Hourly for parcels not yet delivered, with your API key. |
| Tankerkönig | While a fuel widget is open, at most every five minutes. |
| MusicBrainz and Cover Art Archive | Artist search when you use it; releases of followed artists every six hours, at most one request per second. |
| Steam and Epic Games Store | Every twelve hours once game deals are switched on. |
| Launch Library 2 and NASA APOD | Once switched on, only while a space widget shows them, at most hourly and every three hours. |
| Container registries | Image digests of your running containers, at most every six hours per image, only with the Docker update check on. |
| GitHub releases API | At most once a day, only with the update check under Settings → About on. |

## Pages you save

Bookmarks and the read-later list fetch the pages you give them:

- **Bookmark details** (off by default): the page's title and icon, once when
  you add a bookmark, at most 512 KB per page and 64 KB per icon.
- **Link check** (off by default): each bookmarked page once a week.
- **Read later**: the page and its images, once, when you save it.

These requests time out after eight seconds, follow at most five redirects
and never connect to loopback, link-local (including the cloud metadata
address `169.254.169.254`), unspecified or multicast addresses. Addresses in
your LAN stay allowed, because home-server bookmarks point there on purpose.

## Why static analysis complains

Code scanners report several of these requests as "server-side request
forgery", because the server fetches URLs that come from user input. This is
intended: only the single owner can enter those URLs, behind a session, CSRF
protection and the network allow-list. Integration targets are additionally
validated when saved and again on every connection (see
[integration targets](security-model.md#integration-targets-191)).

Two other findings come up regularly:

- **MD5 in the FRITZ!Box login.** TR-064 uses HTTP digest authentication,
  which the protocol defines with MD5. The password is not stored as an MD5
  hash; it is encrypted with `SECRET_KEY`, and MD5 only answers the box's
  login challenge.
- **Open redirect in the search endpoint.** The endpoint redirects to the
  search engine or to a URL typed into the search field. It requires a
  session, and the `SameSite=Strict` cookie is never sent with requests from
  other sites, so it cannot be abused to redirect strangers.
