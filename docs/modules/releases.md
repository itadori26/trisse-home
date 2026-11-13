# Releases

**What it does:** New albums, singles and EPs of the artists you follow, from [MusicBrainz](https://musicbrainz.org), with cover art from the Cover Art Archive.

## Configuration

No environment variables. Follow artists in the widget (artist search runs against MusicBrainz).

## Outbound requests

MusicBrainz: artist search when you use it, and the releases of followed artists every 6 hours, at most one request per second. Cover Art Archive for covers.

## API

`/api/releases/...` (followed artists, upcoming and recent releases) – see `api/openapi.yaml`.
