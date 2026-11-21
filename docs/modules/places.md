# Places

**What it does:** you enter a place once, by address search or with your
browser's location, and choose it in every widget that needs a location.

## Using places

Open **Settings → Connections & services → Places**. There, search for an
address or town, or press *Use my location*, give the place a name such as
"Home" and save it. Each place shows which widgets use it; you can rename or
remove it there. Removing a place that widgets use asks first and names
them. Those widgets keep the location they had, they just no longer follow
the place.

In the settings of the weather and rain radar widgets and in Settings →
Fuel prices, *Saved place* lists your places. Choosing one fills in the
location; the coordinates stay out of the way under *Advanced*. A location
found by the search can be kept with *Save as a place*.

*Use my location* also sits next to *Saved place* in those settings. Your
browser asks for permission the first time; Waylight reads the position only
when you press the button. It rounds the coordinates to three decimals
(about 100 m), saves them as a place called "My location" and picks it. If a
place already exists at that spot, the widget uses that one instead. The
same rounding applies in Settings → Places.

Places belong to the account that saved them. Like other personal content
they are visible to the other accounts of the instance by default, but only
their owner can change them. One account keeps at most 50 places.

The departures widget uses places to find a stop: choose a place under
*Saved place* and the stops nearest to it appear as a list to pick from.
The widget then keeps only the stop. The waste calendar still reads an ICS
calendar and does not take a place yet.

## Outbound requests

None of its own. The address search uses the weather module's geocoder
(Open-Meteo); without the weather module only the browser location is
offered. The browser location never leaves your browser and this server,
but the browser itself may ask its maker's location service to work out
where you are (Firefox and Chrome use one when there is no GPS); that is
between you and your browser. Listing stops near a place asks the journey
planner, see the departures entry in
[outbound requests](../outbound-requests.md).

## API

`GET /api/places` lists the places with the widgets on your own tabs that
use them (`usedBy`); `POST /api/places`, `PUT /api/places/{id}` and
`DELETE /api/places/{id}` change them. A widget names its place with
`placeId` in its settings, next to the coordinates it copied. See
`api/modules/places.yaml`.
