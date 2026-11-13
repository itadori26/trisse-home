# Holidays

**What it does:** The next public holidays of a German state, with weekday and a countdown, and optionally the school holidays (#55). The state is chosen per widget (default: Lower Saxony); current holidays are highlighted.

## Configuration

- **Public holidays** are computed on the server for all 16 states (Easter-dependent holidays, Reformation Day, Buß- und Bettag, Women's Day in Berlin and Mecklenburg-Vorpommern, …). No configuration and no request is needed. Holidays that only apply in some municipalities (Corpus Christi in parts of Saxony and Thuringia, Assumption Day in Bavaria, Augsburger Friedensfest) are not included.
- **School holidays** (optional): Settings → Holidays, the ICS link of your state's school holiday calendar (`https://`, `http://` or `webcal://`), e.g. from your ministry of education or a school holiday website. The link is tested before it is saved. Like every owner-configured address it is checked centrally (#191): no credentials in the link, never link-local or cloud metadata addresses, and redirects are not followed (enter the final address).

## Outbound requests

None for public holidays. With an ICS link: the link when it is saved, then at most once a week (the background job checks daily). The periods are cached per year in the database, so restarts need no request; after a data import they are loaded again.

## API

`GET /api/holidays?state=NI&limit=8`, `GET/PUT /api/holidays/settings` – see `api/modules/holidays.yaml`.
