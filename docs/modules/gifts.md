# Gift ideas

**What it does:** Private gift ideas per person, grouped by person and sorted by the next birthday when the birthdays module is on.

## Using it

Add the *Gift ideas* widget and write down an idea with **+ New idea**: who it
is for, what it is, and optionally a link and a note. Tick an idea off once
it is bought; bought ideas stay listed under the person until you remove
them.

When the birthdays module is on, the widget shows each person's next
birthday next to their name, and the people whose birthday comes first are
listed first. Names are matched as written, so use the same spelling as in
the birthdays widget.

The compact view counts the ideas that are still to buy. It never names a
person, so a compact tile on a shared screen gives nothing away.

## Who sees what

Gift ideas are always private. Only the account that wrote an idea sees,
changes and removes it; there is no sharing option, guests never see ideas,
and admins have no access to other people's ideas either. A kiosk or a shared
tab shows the widget empty for everyone but its author.

Like everything else, ideas are part of the instance backup.

## Configuration

No environment variables. `MODULES_DISABLED=gifts` switches the module off.

## Outbound requests

None; the module only uses the database. Links are opened by the browser.

## API

`/api/gifts` (list, create, update, delete) and `/api/gifts/compact` – see
`api/modules/gifts.yaml`.
