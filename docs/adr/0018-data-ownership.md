# ADR 0018: Data ownership for several full accounts

- Status: Accepted (2026-11-15, by the maintainer)
- Date: 2026-11-15
- Issue: [#293](https://github.com/itadori26/waylight/issues/293), part of [#292](https://github.com/itadori26/waylight/issues/292)
- Extends: [ADR 0012](0012-owner-and-guest-roles.md)

## Context

Today an instance has one owner and up to 20 read-only guests (ADR 0012). All
data belongs to the instance: a note or a to-do has no owner, and a guest who
sees the notes widget on a shared tab can read every note. For a small
organisation (#292) several people need their own dashboards and their own
content, and some content must be shared with a group or with everyone.

The rule for who owns a record touches every module that stores user content.
There are 55 modules today, and the cost of the change grows with each one, so
the decision has to come before the account work in #294–#305.

What already exists and has to keep working:

- `kernel.Principal` travels in the request context (account ID, role, shared
  tabs, token scopes); background jobs run without one and act as the owner.
- Authorization is decided centrally by the auth module's guard, as an
  allow-list; controllers contain no permission checks.
- Guests, kiosk tokens (#50) and API tokens (#96) reuse this model.
- Backups copy the whole SQLite database; the JSON export dumps tables per
  module; config as code (#88) exports instance settings.

## Options

### 1. Ownership per record

Every table with user content gets an `owner_id` and a `visibility`
(`private`, `group:<id>`, `everyone`). Repositories read the principal from
the context and filter: a person sees their own records and those shared
with them.

- Fine-grained sharing (one note for the team, the rest private) without
  copying data.
- One database, one backup, one set of migrations. Shared tabs, groups and
  admin views stay simple joins.
- Risk: a repository that forgets the filter leaks records. This needs a
  shared helper and tests that prove the filter per module (see Decision).
- Every content table needs a migration; existing rows need an owner.

### 2. One database per person plus a shared one

Each account gets its own SQLite file for personal content; instance data
and shared content live in a common file.

- Strict separation by construction: a missing filter cannot leak, because
  the other person's file is not open.
- Shared content is awkward: a note that becomes "for the team" has to move
  between files; a group (not everyone) needs a third kind of file.
- Migrations run per file and per account; a failed migration in one file
  leaves accounts on different schema versions.
- Backups, the JSON export and restore have to handle N files consistently.
- Up to N open SQLite connections and WAL files; harder on a Raspberry Pi.

### 3. Module-level partition without sharing

Every content table gets an `owner_id`, but no visibility: each person only
ever sees their own records; sharing happens only through team tabs that
admins maintain.

- Simplest filter (`owner_id = ?`).
- Does not cover the cases #292 is about: a shared to-do list for the
  office, team notes, a shopping list for the household. They would have to
  become separate "team" modules.

## Decision

**Option 1, ownership per record, applied only where a module stores
personal content.** Instance configuration stays owned by the instance and is
managed by admins.

### Three kinds of module data

| Kind | Owner | Modules (today) |
|---|---|---|
| **Personal content** | the account that created it; visibility per record | notes, todos, bookmarks, habits, countdown, focus, readlater, drop, birthdays, photos, study, shopping, rss (subscriptions), releases, gamedeals, markets (watchlists), parcels, meals |
| **Personal settings and connections** | one row per account | appearance (theme, language, greeting), search (shortcuts, prefixes), briefing, onboarding (tour state), calendar (CalDAV, ICS, Google/Microsoft accounts), mail (IMAP accounts), todos sources (#169–#171) |
| **Instance** | the instance; admins manage it, others read it as their role allows | services, docker, adguard, proxmox, smarthome, system, phonelist, diskhealth, wireguard, speedtest, fritzbox, backup, backupmonitor, timeline, notifications (channels), transfer, device, summary (Ollama), waste, holidays, weather, transit, electricity, fuel, space, media, arr, immich, nextcloud, sports, apiwidget, auth, dashboard templates |

The dashboard is its own case: a tab belongs to an account (its personal
tabs) or is a **team tab** maintained by admins and shown to groups (#295).
Widgets belong to their tab.

A module states its kind in code, next to its migrations (a constant on the
module, checked by a test in `cmd/server`), so a new module has to decide.

### Visibility

- `private` – only the owner. **Default for new records of members.**
- `group:<id>` – the owner and members of that group (#294 introduces groups).
- `everyone` – every account of the instance, guests included when the
  widget is on a tab they may see.

Admins have no automatic read access to other people's private records. An
admin can delete an account (and with it its records) or transfer its records
to another account, but not browse them. This keeps the promise of private
data honest in a small office, where the admin is often a colleague.

### Enforcement

Forgetting a filter is the main risk of option 1, so the filter is not left
to each repository by hand:

- `internal/shared/kernel` gets a `Visibility` value object and
  `VisibleTo(Principal)`, the one rule for "may this principal see this
  record".
- `internal/shared/sqlutil` gets a helper that adds the matching
  `WHERE owner_id = ? OR visibility IN (…)` clause from the principal; the
  repositories of personal-content modules use it for every read.
- `internal/archtest` gets a rule: a repository of a module marked
  *personal content* must not query its content tables without the helper
  (checked on the SQL strings, as other rules already do).
- The end-to-end registry (`cmd/server/api_<module>_test.go`, #194) gains a
  **second member** in its fixture, and each personal-content module's check
  proves that member B cannot read, change or delete a private record of
  member A. A module cannot be marked personal content without this check.
- Writes always set `owner_id` from the principal, never from the request
  body.

### Roles and tokens in the model

- **Admin** (the former owner becomes the first admin): manages accounts,
  groups, team tabs and instance settings; for personal content they are a
  member like everyone else.
- **Member**: own tabs, own personal content, can share records with a group
  or everyone.
- **Guest**: read-only as today. Sees the tabs shared with them and, in those
  widgets, only records with visibility `everyone` or a group the guest is in.
- **Kiosk token**: belongs to the account that created it and shows that
  account's chosen tab read-only, with the records that account sees.
- **API token**: belongs to its account; reads what the account sees, writes
  as the account (`owner_id` = the account), limited by its scopes as today.
- **Background jobs**: run as the system. Jobs that act for a person (to-do
  reminders, the daily briefing push) read the owner of the record and send
  to that person's channels; until per-person channels exist (#303 and
  later), they use the instance channels as today.

### Migration of existing instances

One migration per personal-content module, in the same release:

1. Add `owner_id` (`NOT NULL`, default the existing owner's account ID) and
   `visibility` (`NOT NULL`, default `everyone`) plus an index on `owner_id`.
2. Every existing row therefore belongs to the former owner with visibility
   `everyone`. **Guests see exactly what they saw before**, because ADR 0012
   shared all module data with them anyway.
3. The former owner's **new** records default to `everyone` as long as the
   instance has a single admin and no members, so a household that never
   adds members notices no change. When the first member is added, the
   admin is asked once whether their own new records should default to
   private from then on.
4. Personal settings tables get an `account_id` column; the existing single
   row becomes the owner's.

No data is copied or moved between files, so a failed migration is rolled
back by SQLite's transaction like any other.

### Backups, export and config as code

- Backups stay a copy of the one database and contain everyone's data.
  Restoring is an admin action for the whole instance.
- The JSON export (`transfer`) exports what the requesting account owns;
  admins can additionally export the instance (all accounts).
- Config as code (#88) stays instance-level: team tabs, services and the
  other instance sections. Personal tabs are not part of `waylight.yaml`.

### Core or `ee/`

Ownership, visibility and the filter live in the core (MIT): a household
with two adults needs private notes as much as an office does, and ADR 0017
already places shared foundations in the core. Group management at scale,
team tabs maintained by admins, SSO group mapping, audit log and admin
policies are team features and go to `ee/` as decided in their issues.

## Consequences

- Every personal-content module needs a migration, the helper in its
  repository, and the second-member check in its end-to-end test. That is
  the follow-up work below; it can be done module by module, because a
  module without ownership keeps today's behaviour (everything `everyone`).
- Queries of personal content get one extra indexed condition. With SQLite
  and the expected data sizes (thousands of records) this is not measurable.
- Guests keep their current view after the upgrade; only records created
  private by members are hidden from them.
- An admin cannot read members' private records. Support cases ("I lost my
  note") are solved by the member, a backup restore, or by transferring an
  account's records, not by an admin view.
- Option 2 stays possible later for a module with strict separation needs
  (for example a future password vault), because the kinds of module data
  are explicit.

## Follow-up work

1. `kernel.Visibility`, `sqlutil` filter helper, archtest rule, second member
   in the e2e fixture (core).
2. Roles admin/member/guest and groups in the auth module (#294).
3. Personal tabs and team tabs in the dashboard (#295).
4. Per module, in this order (most used first): notes, todos, bookmarks,
   shopping, readlater, countdown, habits, focus, birthdays, photos, drop,
   study, rss, releases, gamedeals, markets, parcels.
5. Personal settings per account: appearance, search, briefing, onboarding,
   calendar, mail, todos sources.
6. Sharing UI: a visibility control on records of personal-content widgets
   ("Only me", group, "Everyone").
7. JSON export per account and for the instance (`transfer`).
8. Documentation: `docs/accounts.md` (roles, who sees what), and the module
   checklist in `docs/adding-a-module.md` (declare the kind of data).
