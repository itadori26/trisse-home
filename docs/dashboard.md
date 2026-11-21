# Tabs and widgets

Everything you see in Waylight is a widget on a tab. This page explains
how to arrange them, how the phone layout works and how to make a tab open
by itself. For a guided first run, see [first steps](first-steps.md).

## Edit mode

Press `e`, or choose *Edit* in the account menu, to change the dashboard.
Edit mode shows a frame around each widget with buttons for its settings,
duplicating and removing it, and the tab bar gains controls for adding,
renaming, moving and deleting tabs. Press `e` again or choose *Done* to
leave. There is no save button; each change is stored as soon as you make
it.

`Ctrl+Z` undoes the last change in edit mode and `Ctrl+Shift+Z` or `Ctrl+Y`
redoes it. A removed widget comes back with its settings and its content.

## Adding and arranging widgets

**+ Add widget** opens the picker. Widgets are grouped into *Everyday*,
*Information*, *Server* and *Network & integrations*. Type to filter, use the
arrow keys to move through the list and `Enter` to add it to the current
tab.

A click on a widget adds it in its default size. The ⊞ button in the corner
of each entry shows the widget in every [size](#widget-sizes) side by side
instead, with example data, so you can see what a small tile or a strip will
say before the widget is set up. Choosing a preview adds the widget in that
size.

The grid has twelve columns. Drag a widget by its frame to move it and drag
its lower right corner to resize it; other widgets move out of the way. With
the keyboard, focus a widget and use the arrow keys to move it or `Shift`
with the arrow keys to resize it. A screen reader announces the new
position.

## Widget sizes

In edit mode every widget has a size menu in its title bar with the same
sizes as the phone layout: *Small*, *Medium*, *Large* and *Strip*. The size
decides what the widget shows. Medium and large tiles show the full widget.
Small tiles and strips show its compact view, one or two lines such as
*3 open · 1 overdue*, and a click on the tile opens the full widget in a
panel at the side of the screen (a sheet from the bottom on phones). `Escape`
or the close button closes the panel and puts the focus back on the tile.
Below the widget the panel offers *Edit widget* when you may change the tab
and the widget has settings, and for widgets that watch something that can
fail (services, backups, drives) admins get *Notify me when this fails*,
which opens the notification settings. Kiosk displays show the compact view
without the panel.

When a narrow window makes a medium or large tile too small for its full
view, the widget switches to its compact view instead of cutting its content
off.

Dragging the corner of a sized widget snaps to the nearest size; with the
keyboard, `Shift` and the arrow keys step through the sizes. *Free size* in
the menu turns a widget back into a freely sized tile, which always shows the
full widget. Widgets placed before sizes existed stay free until you pick a
size, so existing dashboards keep their look.

Some compact views carry one quick action next to the short state, so the
most common step needs no panel: habits tick off the next habit due today
and to-dos tick off the most urgent one (overdue first, then the earliest due
date). A smart home tile with one device shows its state and turns it on or
off (or runs a button, script or scene) when actions are enabled for it with
`SMARTHOME_ACTIONS_ENTITIES`. The focus timer offers **Start** or **Pause**
and shows the running phase as *Focus · 12:40*; the timer is kept per
account on the server, so the full widget, the compact view and your other
devices show the same one. The button gives feedback at once; when the server refuses, its
message is shown and nothing changes. Guests and kiosk displays do not see
quick actions.

Compact views exist for services and the status overview (*11 of 12
running*, and which service is down), the backup status, drive health, the
calendar (the next event), to-dos (the built-in list), habits, the focus
timer, smart home tiles with one device, Docker (running containers), parcels
(on the way and arriving today), the speed test (the latest measurement),
Proxmox (running guests), the DNS filter (share of blocked queries),
WireGuard (connected peers), the weather (temperature and sky at the
widget's location), electricity prices (the current price and when the
cheapest hours begin) and, with the `ee` features, presence (with chips for
your own status).
Widgets whose content is the point, such as the clock, bookmarks or
notes, show their full view at every size.

You can put the same kind of widget on a tab more than once, for example two
weather widgets for two towns. **Duplicate** copies a widget with its
settings.

## Tabs

A tab has a name, an optional background image and its own set of widgets.
In edit mode you can:

- add a tab with **+ Tab**, either empty or from a template;
- save a tab as a template, so you can stamp out similar tabs later
  (templates keep the layout and widget settings, not the content such as
  notes or to-dos);
- choose a background for one tab, which overrides the background of all
  tabs (see [background image](appearance.md#background-image));
- move tabs left and right, rename and delete them.

`1` to `9` switch to the first nine tabs.

## Templates

On the first sign-in, while the dashboard is still empty or holds only the
example tabs of the sample data, Waylight asks **What do you use Waylight
for?** and offers four templates, each with a small preview of its layout and
the widgets it contains. The guided tour starts once the question is
answered.

| Template | Today | My things | System (admins) |
| --- | --- | --- | --- |
| Family | Weather, to-dos, calendar strip, waste collection | Shopping list, birthdays, notes | Services strip |
| Homelab | Clock, to-dos | Bookmarks | Services strip, containers, backups, drive health, host metrics |
| Work | Calendar strip, to-dos, focus timer | Mail, notes, bookmarks | Services strip |
| Wall tablet | Large clock, weather, calendar | Departures | Services strip |

A template creates a tab laid out in [zones](#zones) with its widgets
placed, including the phone layout. Widgets with a compact view start small
or as a strip; status widgets in the System zone feed the attention line.
Templates with a weather widget ask for a location first (search or *Use my
location*; *Set it up later* skips it). Widgets that need a service or a
setting show *Set up* or what is missing. Kiosk mode and night dimming for a
wall tablet are set on the kiosk display itself.

With sample data, a template replaces the example tabs; the example to-dos,
notes, bookmarks and services stay. *Keep the example dashboard* keeps the
example tabs as ordinary tabs, and so does closing the dialog. *Start with
an empty dashboard* removes them (with sample data) or just closes the
question.

The same templates appear when you create a tab (**+ Tab**), above your own
saved templates. They are YAML files in the format of the `dashboard`
section of [waylight.yaml](config-as-code.md), one tab each, under
`internal/modules/dashboard/adapter/configfile/templates/`; the tests check
that each one parses, places its widgets without overlap, puts every widget
in a zone and gives sized widgets the cells of their size.

## Opening a tab by default

Normally the dashboard opens the first tab. With **Open by default** you can
give any tab rules for when it should open instead. A rule combines
weekdays, a time window and client networks, and each part is optional:

| Rule | Opens the tab … |
| --- | --- |
| Mon–Fri, 06:00–09:30 | on weekday mornings |
| 192.168.10.0/24 | when the browser is in that network, for example at work over VPN |
| Sat–Sun | at weekends, all day |

A tab can have several rules; any one of them is enough. When several tabs
match, the first tab in the tab bar wins. Rules use the server's time zone
(`TZ`). Network rules see your real address only if Waylight trusts your
reverse proxy (`TRUSTED_PROXIES`, see [configuration](configuration.md)).

Right after you create a tab, a short question above the grid offers three
quick choices: mornings (06:00–10:00), weekdays from 08:00 to 17:00, and "in
this network". The last one only appears when your browser is in a private
network; it uses the /24 (IPv4) or /64 (IPv6) around your address. Each
choice becomes an ordinary rule you can change later under **Open by
default**. If you answer **No, thanks** or reload the page, the question does
not come back for that tab.

## The phone layout

On narrow screens the widgets are stacked in one column. By default the
order follows the desktop layout, row by row. Each tab can have its own
phone layout instead: in edit mode on a phone, move widgets up and down, make
them shorter or taller, or hide them on phones only. **Reset from desktop
layout** discards the phone layout of that tab.

Give a widget a size to place it on a grid of four columns, as on a phone's
home screen: **Small** (two by two, two side by side), **Medium** (full width,
two cells high), **Large** (four by four) or **Strip** (one cell high). Widgets
without a size keep the full width and their height. The mobile app shows the
same grid.

The phone layout belongs to the tab, not to the device, so it applies to
every phone you use.

## The attention line and the system summary

When something on a tab needs a look, one line at the top says what it is,
for example *Services: 11 of 12 running · Nextcloud is down*, and how many
other things need a look as well. **View** scrolls to the widget and puts the
focus on it. The line takes the status widgets of the tab whose short state
is amber or red (services, backups, drives) and connections that stopped
working, such as an expired sign-in. It disappears when all is well. Colour
is used only for what needs action: amber asks for a look, red means
something is broken.

A tab with two or more status widgets also gets a summary line, for example
*Services: 12 of 12 running · Backups: Backup OK · Disk health: 3
drives healthy*. **Collapse** folds the status widgets into that line and
**Show all** brings them back. The choice is remembered per tab in this
browser; in edit mode all widgets are shown.

## Zones

A tab can be laid out in three zones instead of the free grid: **Today**
(clock, weather, calendar, waste collection, to-dos), **My things** and
**System** (server and network widgets). Switch them on in edit mode with
**Zones** in the bar at the bottom; the setting is stored with the tab.

Each widget falls into a zone by its kind. To put it somewhere else, pick
another zone in the menu next to its size in edit mode. Within a zone the
widgets keep their size and follow the order they had in the free grid, so
switching zones off again brings back the old layout unchanged. On phones the
zones stack, one widget below the other.

Only admins see the System zone; members and guests see Today and My things.
Outside edit mode a zoned tab stays quiet: widgets with nothing to show, such
as a calendar without upcoming events or a to-do list with nothing open, are
left out until there is something to show again.

## Widget status

Widgets that load data from the server refresh on their own, at an interval
that suits the source. They pause while the browser tab is in the background
or the widget is scrolled out of view. Hover over a widget or move the keyboard focus into
it to see when its data was last updated and to refresh it by hand. When the data is older than expected or the last update failed, the
badge stays visible and changes colour, so a broken integration does not
look like current data.

Some widgets also have a short state: one line such as *3 open · 1 overdue*
for to-dos, *2 of 5 today* for habits or *50 min focus today* for the focus
timer, with a level (none, ok, attention, problem). Small tiles and strips
on the dashboard ([widget sizes](#widget-sizes)), the app's home screen
widgets and the watch show it instead of the full widget; it is
served at `GET /api/<module>/compact` with the same permissions as the
module's other endpoints. `GET /api/modules` lists under `compact` which
modules you may use have one, with the path and the query parameters it
takes (the smart home device needs `?entity=`), so clients do not need a list
of their own.

## Setting up a widget in place

A widget whose service is not connected yet shows one sentence and a
**Set up** button. The setup runs inside the widget in three steps:

1. **Address.** Enter where the service runs, for example
   `http://192.168.178.20:2283` for Immich. For admins Waylight first looks
   for the service in the home network and offers what it found
   ("Immich found at http://172.18.0.5:2283"); entering an address by hand
   always works.
2. **Credential.** The widget lists the steps in the service's own words,
   for example *Account Settings → API Keys → New API Key* in Immich, and
   *Open Immich* takes you to that page in a new tab. Paste the key or token.
3. **Check.** Waylight tries the connection before it saves anything. When
   it works you see real data (the number of photos and three of today's
   memories, a few entities of Home Assistant, the calendars found and your
   next events) and press *Done*.

When the check fails, the message says why and what to try: the key was not
accepted, nothing answers at the address, something other than the expected
service answers there, or the certificate is not trusted.

### Signing in instead of pasting a key

For Nextcloud (the Nextcloud widget, the calendar and a shopping list on
Nextcloud) and for Home Assistant the credential step starts with **Sign in
with Nextcloud** or **Sign in with Home Assistant**. Pressing it opens the
service's sign-in page in a new window. You sign in there and allow access,
then come back: the assistant says when it has the access and checks the
connection as usual. Nothing has to be copied, and the key never passes
through your browser. *Paste a key instead* switches to the steps above at
any time, and the assistant switches by itself when signing in cannot start.

Nextcloud uses its [Login Flow v2](https://docs.nextcloud.com/server/latest/developer_manual/client_apis/LoginFlow/).
Waylight starts the flow on the Nextcloud address you entered and asks that
address every two seconds whether you have granted access, for at most ten
minutes or until you close the assistant. Nextcloud then creates an app
password named *Waylight*, which you can revoke under *Settings → Security →
Devices & sessions*. For a shopping list the Nextcloud address is taken from
the list address (the part before `/remote.php/`); other CalDAV servers keep
the pasted password.

Home Assistant uses OAuth with Waylight's own address as the client ID, so
nothing has to be registered in Home Assistant. That address comes from
`PUBLIC_URL`: Home Assistant sends your browser back to
`PUBLIC_URL/api/setup/signin/oauth/callback`, so the browser that signs in
must be able to open it. Without `PUBLIC_URL` the assistant explains this
and asks for a long-lived access token. After signing in, Waylight keeps a
refresh token, encrypted like any other key, and renews the short-lived
access token by itself. Home Assistant lists the sign-in under your profile
as a refresh token for Waylight's address; deleting it there disconnects
the widget until you set it up again.

A sign-in lives in memory only, for ten minutes, and belongs to the account
that started it. The link back from Home Assistant carries a random value
that works once; a sign-in that was not finished, or that another account
started, cannot be used.

The assistant exists for Immich, Home Assistant, the Nextcloud calendar,
Nextcloud, Grafana, Proxmox VE, Sonarr, Radarr and Prowlarr, Jellyfin and
Plex, AdGuard Home, Pi-hole, Speedtest Tracker, the FRITZ!Box, fuel prices
and a CalDAV task list for the shopping list. Where one widget
covers several products, the address step asks which one runs there. Proxmox
usually runs with a self-signed certificate; pin it under Settings → Proxmox
VE, the assistant does not ask for the fingerprint. Speedtest thresholds stay
off until you set them under Settings → Speedtest Tracker. The FRITZ!Box internet
widget works without a FRITZ!Box user (status over UPnP); the call list
needs one. Fuel prices ask for a place instead of an address and find it
with the place search of the weather module; radius (5 km) and fuel (E10)
are changed in the settings. The shopping list offers the CalDAV assistant
next to creating a list on Waylight, for every account that is not a guest. For the calendar you enter the Nextcloud address, your user name and
an app password; Waylight finds your calendars itself. Immich and Home
Assistant are set up for the whole instance, so only admins see *Set up*
there; every account can connect its own calendar. The settings pages stay
the place to change a connection later.

### How services are found

Waylight looks only when a setup opens, never in the background. It checks
Docker containers whose image belongs to the service (when the docker
module can read the Docker API), services that announce themselves in the
local network by mDNS / DNS-SD (Home Assistant as `_home-assistant._tcp`),
the names the service usually has in a home network such as
`homeassistant.local`, and the service's default port on the gateway. Only private addresses (RFC 1918, ULA) that `ALLOWED_CIDRS`
also allows are tried, on a fixed port per service and at most twelve
requests per search; ports are never scanned. Each address must answer like
the service (Immich's ping, Home Assistant's manifest, Nextcloud's
`status.php`) before it is offered. Results are kept for three minutes and
not stored.

For mDNS Waylight sends one query per service type and search over IPv4
and collects answers for about a second; it does not listen on port 5353
or announce anything. Multicast only reaches Waylight when its container
uses the host network (`network_mode: host`) or runs outside Docker; in a
bridge network the other sources still work.

Besides Immich, Home Assistant and Nextcloud, the setup finds the
FRITZ!Box (its TR-064 description on port 49000), Jellyfin (its public
system info), Grafana (its health check) and AdGuard Home (its login page).
Plex, Pi-hole, Proxmox VE, the *arr apps and Speedtest Tracker are entered
by hand.

## What needs your attention

Waylight keeps a list of what needs your attention: widgets on your own tabs
that miss a setting (a departures widget without a stop) and services behind
your widgets that are not set up or stopped answering (Immich, Home
Assistant, Nextcloud, Grafana, the FRITZ!Box, Proxmox, the media server,
Sonarr/Radarr/Prowlarr and Speedtest Tracker for now). Admins also see
broken connections of the instance that none of their widgets use. Shared
tabs of others and integrations you never set up are left out. Each entry
names one thing to do, such as *choose a stop* or *check again*.

In the web interface a dot with the number of entries appears in the top
bar as soon as something needs you; it is hidden while everything works.
It opens **Settings → Needs your attention**, which lists what stopped
working first and what is not set up yet below. The button next to an entry
opens the widget's own settings on its tab, opens the integration's
settings, or asks the service again. The list updates by itself while the
page is open, so a fixed entry disappears without reloading.

The app reads the same list from `GET /api/attention` (see the API
reference). A service's state is kept for two minutes; *Check again* and
closing the settings after a change ask the services right away
(`?refresh=true`).

## Services and places

**Settings → Connections & services** has two pages. *Connected services* lists every
service your widgets connect to with its state: connected, with a problem
(and the cause), or not set up yet. Each entry names the widgets that use it
and opens the service's settings; *Check all again* asks every service right
away. Admins see all services of the instance; other accounts see their own
connections and the instance services their widgets depend on. *Places*
holds the places you entered once for weather, rain radar and fuel prices
(see [Places](modules/places.md)).

## Related pages

- [Feature overview](features.md): every widget and what it can do.
- [Accounts, guests and kiosk displays](accounts.md): share tabs with others.
- [Accessibility](accessibility.md): all keyboard shortcuts.
