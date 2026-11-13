# Habits

**What it does:** Habit tracker widget: daily habits or chosen weekdays, one tap to check off today (or yesterday), current and best streak, and an optional 4-week heatmap. Days off do not break a streak.

## Configuration

No environment variables. Habits are created in the widget itself; the widget settings choose whether the heatmap is shown.

## Outbound requests

None; the module only uses the database.

## API

`/api/habits` (list, create, update, delete) and check-offs for today and yesterday – see `api/openapi.yaml`.
