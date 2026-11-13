# Fuel prices

**What it does:** Fuel prices in Germany from [Tankerkönig](https://creativecommons.tankerkoenig.de) (MTS-K data, CC BY 4.0): the cheapest open stations around a place for E5, E10 or Diesel, or your favourite stations.

## Configuration

Settings → Fuel prices: a free Tankerkönig API key (stored encrypted), the location, radius and fuel type. `FUEL_API_URL` changes the API (only for a mirror or tests).

## Outbound requests

Only the Tankerkönig API, while a fuel widget is open, at most every 5 minutes. The API key is sent with each request.

## API

`GET/PUT /api/fuel/settings`, `GET /api/fuel/overview` and favourites – see `api/modules/fuel.yaml`.
