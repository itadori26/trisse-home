# Game deals

**What it does:** Discounts on your public Steam wishlist (price, discount, end date, minimum discount) and the current free games of the Epic Games Store.

## Configuration

No environment variables. Enter your public Steam profile in the widget settings and choose a minimum discount.

## Outbound requests

The Steam and Epic Games stores, every 12 hours once enabled.

## API

`/api/gamedeals/...` – see `api/modules/gamedeals.yaml`.
