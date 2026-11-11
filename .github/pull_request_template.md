## Linked issue

Closes #

## Changes

-

## Checklist

- [ ] The change respects the dependency rule (domain → use case → adapter → platform)
- [ ] Tests for every affected ring (domain, use case, gateway, controller, frontend)
- [ ] `make ci` passes locally (and `make smoke` for UI changes)
- [ ] Docs updated where needed (README, `api/openapi.yaml`, `.env.example`, ADR)
- [ ] Generated TypeScript types are up to date (`make generate`)
- [ ] No new runtime dependency, or it is justified below (size, maintenance, why not stdlib/own code); resource budgets green ([ADR 0011](../docs/adr/0011-lean-by-default.md))
