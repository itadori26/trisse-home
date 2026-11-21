# Feature overview

Every feature of Waylight is a module. Modules you do not need can be
switched off with `MODULES_DISABLED` (see [configuration](configuration.md));
their widgets, settings and background jobs then disappear completely. The
module name is given in brackets.

Modules marked *opt-in* make no outside request and show nothing until you
set them up. Pages under [modules](modules/README.md) describe the newer
modules in detail.

## Everyday

| Module | What it does |
| --- | --- |
| To-dos (`todos`) | Due dates with overdue highlighting, recurring to-dos (daily, on chosen weekdays, monthly, or N days after completion) and reminders through your notification channels. Works offline. A widget can show a connected CalDAV task list (Nextcloud Tasks, Synology, Radicale, …) a Todoist project or a list in Microsoft To Do, Google Tasks or TickTick instead ([details](modules/todos.md)). |
| Notes (`notes`) | Markdown notes that can be pinned to the dashboard, with full-text search, `[[wiki links]]` and backlinks, and a version history with diff and restore. |
| Calendar (`calendar`) | CalDAV calendars such as Nextcloud, including recurring events, Google Calendar and Microsoft 365 accounts, plus read-only ICS subscriptions. Agenda, month or split view with event details; new events can be added from the widget. [Details](modules/calendar.md) |
| Shopping list (`shopping`) | Lists kept on Waylight, shareable with a group or everyone and exportable as text or `.ics`, or a CalDAV task list the household already uses in its task app. See [shopping](modules/shopping.md). |
| Phone list (`phonelist`) | A searchable directory of names, roles, teams, phone numbers and e-mail addresses, maintained by admins; the search field answers "phone <name>". [Details](modules/phonelist.md) |
| Countdowns (`countdown`) | Days left until exams, holidays or deadlines, highlighted when the date is close. |
| Focus timer (`focus`) | Pomodoro timer kept per account on the server, so every device and the compact view show the same timer (in the extension's local mode it stays in the browser), with an optional browser notification and a count of finished sessions per day. |
| Habits (`habits`) | Daily habits or habits on chosen weekdays, checked off with one tap, with current and best streak and a four-week heatmap. |
| Read later (`readlater`) | Saves a clean reader copy of a page with its images, so it stays readable when the original disappears. [Details](modules/readlater.md) |
| Drop zone (`drop`) | Moves text, links and files up to 25 MB between your devices; items expire after one hour, a day or a week. [Details](modules/drop.md) |
| Photo frame (`photos`) | Your own pictures one after another, uploaded to Waylight and scaled down there; albums per widget, interval and fit. In backups, not in data exports. [Details](modules/photos.md) |
| Study progress (`study`) | Modules with credits, semester and grade, progress towards the degree and a weighted average. [Details](modules/study.md) |
| Bookmarks (`bookmarks`) | Links in folders with tags. Opt-in: fetch title and icon when adding, and a weekly check for broken links. Browser bookmarks can be imported. |

## Information

| Module | What it does |
| --- | --- |
| Weather (`weather`) | Forecast from Open-Meteo, no API key needed. In Germany also official DWD warnings, pollen forecast, UV index and a rain radar widget. |
| Places (`places`) | Places entered once by address search or browser location and chosen by the weather, rain radar and fuel price settings; Settings shows which widgets use each one. [Details](modules/places.md) |
| Integrations (`integrations`) | Named connections to a service with an encrypted credential, saved once by an admin and later chosen by every widget of that service; lists which widgets use each one. [Details](modules/integrations.md) |
| Clock (`clock`) | Time and date, optionally in another time zone. |
| Daily briefing (`briefing`) | Weather warnings, failures since last evening, offline services, due to-dos, appointments and waste collection in one widget, optionally pushed at a set time each morning. |
| Feeds (`rss`) | RSS, Atom and JSON Feed with read state, OPML import and export, and a built-in catalogue of news sources by topic and language (Settings → Feeds → Browse catalogue; no outside request until you subscribe; [format](https://github.com/Waylight-App/waylight/blob/main/web/src/features/rss/catalogue/README.md)). |
| Public transport (`transit`) | Next departures of a stop with real-time delays, via [Transitous](https://transitous.org). |
| Waste collection (`waste`) | Collection dates from your municipality's ICS calendar, by link or uploaded file; tomorrow's bin is highlighted. |
| Holidays (`holidays`) | Public holidays of a German state, computed locally, and school holidays from an ICS link. [Details](modules/holidays.md) |
| Mail (`mail`) | Unread count and the newest unread messages of one or more IMAP inboxes; read-only, headers only. [Details](modules/mail.md) |
| Parcels (`parcels`) | DHL tracking with your free API key; Hermes, DPD, GLS and UPS with a link to the carrier. Opt-in: find parcels in shipping e-mails ([guide](parcels-mail.md)). |
| Fuel prices (`fuel`) | Cheapest open stations around a place, or your favourites, in Germany via Tankerkönig. [Details](modules/fuel.md) |
| Markets (`markets`) | Watchlist of currencies (ECB), shares, indices and crypto (Stooq) with daily change and sparkline; no API key. [Details](modules/markets.md) |
| Sports (`sports`) | Last result, next fixture and table position of a followed team, or the top of the table (German football leagues via OpenLigaDB, no API key). [Details](modules/sports.md) |
| Birthdays (`birthdays`) | Next birthdays with age, entered by hand, imported from a vCard file or read from a CardDAV address book; optional reminder and briefing line. [Details](modules/birthdays.md) |
| Electricity prices (`electricity`) | Day-ahead market price per hour for today and tomorrow, cheapest hours highlighted (Germany/Luxembourg, Austria; no API key). [Details](modules/electricity.md) |
| Space (`space`) | Next rocket launches with countdown and NASA's Astronomy Picture of the Day. Opt-in. [Details](modules/space.md) |
| AI summaries (`summary`) | Summarises notes and feed articles with a local Ollama; nothing leaves your network. Opt-in. [Details](modules/summary.md) |

## Home server

| Module | What it does |
| --- | --- |
| Services (`services`) | Tiles for your services with online status, response time, TLS certificate warnings and Wake-on-LAN; a status page with uptime and incidents. Tiles can be created from [Docker labels](host-integrations.md#docker-autodiscovery). |
| System (`system`) | CPU, memory, swap, disks, load, uptime and temperature with a 24-hour history, and a forecast of when a disk will be full. |
| Docker (`docker`) | Container status and health through a read-only proxy. Opt-in: image update hints, and restart and logs for [chosen containers](host-integrations.md#docker-actions). |
| Disk health (`diskhealth`) | S.M.A.R.T. verdict, temperature, sector counts and SSD wear from a [host export](host-integrations.md#disk-health). |
| Backup monitoring (`backupmonitor`) | Your own backup and cron jobs report to a secret URL; missed or failed runs raise an alert ([details](notifications.md#backup-and-job-monitoring)). |
| Timeline (`timeline`) | A log of what happened when: outages and recoveries, failed backups, certificates, disks and security events. |
| Proxmox VE (`proxmox`) | Node load and the state of VMs and containers via a read-only token. Opt-in: start and shut down [chosen guests](host-integrations.md#proxmox-ve). |
| JSON widget (`apiwidget`) | Shows values from any JSON endpoint, picked with JSONPath, with units and thresholds. [Details](modules/apiwidget.md) |

## Network and integrations

| Module | What it does |
| --- | --- |
| DNS filter (`adguard`) | Statistics from AdGuard Home or Pi-hole 6+. Opt-in: pause blocking for 5, 30 or 60 minutes. |
| WireGuard (`wireguard`) | Peers with name, last handshake and traffic, from a [host export](host-integrations.md#wireguard) without keys. |
| FRITZ!Box (`fritzbox`) | Two widgets: internet (status, uptime, external IP, throughput) and calls (missed, incoming, outgoing). |
| Home Assistant (`smarthome`) | Entities as tiles with state and last change. Opt-in: switches and buttons for [chosen entities](host-integrations.md#home-assistant). |
| Nextcloud (`nextcloud`) | Version, your quota and, for admins, free space, users and app updates. |
| Speedtest (`speedtest`) | Results from an existing [Speedtest Tracker](https://github.com/alexjustesen/speedtest-tracker) with sparklines and an alert below a threshold. |
| Sonarr, Radarr, Prowlarr (`arr`) | Upcoming episodes and films, the download queue and health warnings. [Details](modules/arr.md) |
| Jellyfin, Plex (`media`) | What is playing now and what was added recently, with posters. [Details](modules/media.md) |
| Immich (`immich`) | Library size and today's "on this day" memories. [Details](modules/immich.md) |
| Music releases (`releases`) | New albums and singles of artists you follow, from MusicBrainz. [Details](modules/releases.md) |
| Game deals (`gamedeals`) | Discounts on your Steam wishlist and free games on the Epic Games Store. [Details](modules/gamedeals.md) |

## Platform

These modules provide the frame around the widgets. `auth` and `dashboard`
cannot be switched off.

| Module | What it does |
| --- | --- |
| Accounts (`auth`) | Owner account with two-factor authentication and passkeys, read-only guests, kiosk displays, API tokens, single sign-on and a security log. See [accounts](accounts.md). |
| Dashboard (`dashboard`) | Tabs, the widget grid, templates, the phone layout and default-tab rules. See [tabs and widgets](dashboard.md). |
| Search (`search`) | Web search, quick search, answers about your server, calculator and quick capture. See [search](search.md). |
| Appearance (`appearance`) | Language, light and dark mode, theme gallery, backgrounds and custom CSS. See [appearance](appearance.md). |
| Notifications (`notifications`) | ntfy, Gotify, webhooks, e-mail, Slack and Microsoft Teams with importance levels, quiet hours and maintenance windows. See [notifications](notifications.md). |
| Wiki search (`wiki`) | Pages of BookStack, Outline and Wiki.js in the search field, asked while you type; nothing is stored. [Details](modules/wiki.md) |
| Backups (`backup`) | Automatic database backups with a traffic light on the settings start page and as a widget, off-site copies over WebDAV and a guided restore. See [backups](backups.md). |
| Import and export (`transfer`) | All your data as one JSON file, without secrets. See [backups](backups.md#export-and-import). Settings (appearance, search, services, bookmarks, dashboard) also as a readable `waylight.yaml` for Git, with `${ENV}` references and a JSON Schema, and an import from Homepage and Homarr 0.x ([details](config-as-code.md)). |
| Guided tour (`onboarding`) | A short tour after the first sign-in. |
| Device (`device`) | Raspberry Pi image only: run the Pi as a server or also show the dashboard full-screen on a connected display. [Details](modules/device.md) |
