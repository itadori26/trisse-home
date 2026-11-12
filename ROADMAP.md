# Roadmap

trisse-home is planned in public. Every item is a GitHub issue; milestones
group them into releases. The order inside a milestone is a guideline, not a
promise — contributions can move anything forward.

| Milestone | Theme | Status |
| --- | --- | --- |
| [v1.0 – Stable & safe](https://github.com/itadori26/trisse-home/milestone/1) | Release pipeline, vulnerability scanning, security model, backups off-site, security event log, disk health, notification rules, resource budgets | in progress |
| [v1.1 – Everyday UX](https://github.com/itadori26/trisse-home/milestone/2) | Quick capture and actions, recurring to-dos, notes with links and history, mobile layout per tab, undo/redo, keyboard shortcuts, single sign-on, imports from other dashboards, accessibility | planned |
| [v1.2 – Widgets & insights](https://github.com/itadori26/trisse-home/milestone/3) | Generic JSON widget, uptime statistics, event timeline, Pi-hole and Proxmox, more notification channels, config as code, API tokens, guest accounts | planned |
| [Backlog](https://github.com/itadori26/trisse-home/milestone/4) | Ideas without a date: more integrations (see [#91](https://github.com/itadori26/trisse-home/issues/91)), kiosk mode, passkeys, local AI summaries, fun widgets, Raspberry Pi appliance image ([#130](https://github.com/itadori26/trisse-home/issues/130)) | ideas |

- **What changed:** see the [releases and changelog](https://github.com/itadori26/trisse-home/releases),
  generated from Conventional Commits for every version.
- **Want to help?** Issues labelled
  [good first issue](https://github.com/itadori26/trisse-home/labels/good%20first%20issue)
  list the files and tests to touch. Start with [CONTRIBUTING.md](CONTRIBUTING.md).
- **Have an idea or a question?** Open an issue with the feature form or ask in
  [Discussions](https://github.com/itadori26/trisse-home/discussions).

## Principles that do not change

1. **Private by default:** no telemetry, no outbound request you did not configure.
2. **Least privilege:** read-only access first; anything that can change a system is a separate, optional step ([security model](docs/security-model.md)).
3. **Lean:** one binary, SQLite, small images; new dependencies need a reason.
4. **Clean Architecture:** every feature is a module with the same four rings, enforced in CI.
