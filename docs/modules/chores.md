# Chores

**What it does:** Recurring household tasks with the people who take turns, a done button and a compact view that says how many tasks are due today.

## Using it

Add the *Chores* widget and create a chore with **+ New chore**. A chore
repeats every few days (1 is daily, 7 is weekly) and starts on the day you
choose. Give it the names of the people who take turns, one per line. Names
are plain text, so children without an account can be on the list too.

When a chore is done, tick it off. It then moves on by its interval and
passes to the next name on the list. A chore that was overdue counts its
next round from the day it was done, so it does not come back at once. The
widget remembers who did it last.

The compact view shows how many chores are due today, turns to attention
when one is overdue, and offers a button that ticks off the first one.

## Who sees what

Chores are shared with everyone on the instance by default. When you create
a chore you can narrow it to one of your groups or to yourself. Everyone who
sees a chore may tick it off; only the person who created it can change or
remove it. Guests only read.

## Configuration

No environment variables. `MODULES_DISABLED=chores` switches the module off.

## Outbound requests

None; the module only uses the database.

## API

`/api/chores` (list, create, update, delete), `/api/chores/{id}/done` and
`/api/chores/compact` – see `api/modules/chores.yaml`.
