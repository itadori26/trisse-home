import { describe, expect, it } from 'vitest';
import { definitionFor, widgetDefinitions } from './widgets';

describe('widget registry (#116)', () => {
  it('loads every view and settings component lazily and caches the chunk', async () => {
    for (const def of widgetDefinitions) {
      for (const r of [def.view, def.settings]) {
        if (!r) continue;
        const first = await r.load();
        expect(first, def.type.kind).toBeTypeOf('function');
        expect(await r.load()).toBe(first);
      }
    }
  });

  it('has unique kinds', () => {
    const kinds = widgetDefinitions.map((d) => d.type.kind);
    expect(new Set(kinds).size).toBe(kinds.length);
  });

  it('declares compact views for the widgets with a compact state (#595, #594)', () => {
    const modules = Object.fromEntries(
      widgetDefinitions.flatMap((d) => {
        const module = d.compact?.({});
        return module ? [[d.type.kind, module]] : [];
      }),
    );
    expect(modules).toEqual({
      todos: 'todos',
      focus: 'focus',
      habits: 'habits',
      services: 'services',
      'status-overview': 'services',
      calendar: 'calendar',
      'backup-status': 'backupmonitor',
      'backup-health': 'backup',
      diskhealth: 'diskhealth',
      docker: 'docker',
      parcels: 'parcels',
      speedtest: 'speedtest',
      proxmox: 'proxmox',
      adguard: 'adguard?provider=adguard',
      wireguard: 'wireguard',
      electricity: 'electricity?zone=DE-LU',
    });
    // A connected to-do list has no compact state on the server.
    expect(definitionFor('todos')?.compact?.({ source: 'caldav-1' })).toBeUndefined();
    // A smart home tile has a compact view with exactly one device (#614).
    const smarthome = definitionFor('smarthome')?.compact;
    expect(smarthome?.({ entities: ['switch.coffee'] })).toBe('smarthome?entity=switch.coffee');
    expect(smarthome?.({ entities: ['switch.coffee', 'light.desk'] })).toBeUndefined();
    // Weather, DNS filter and electricity prices pass their configuration (#641).
    const weather = definitionFor('weather')?.compact;
    expect(weather?.({ lat: 53.14, lon: 8.21 })).toBe('weather?lat=53.14&lon=8.21');
    expect(definitionFor('adguard')?.compact?.({ provider: 'pihole' })).toBe('adguard?provider=pihole');
    expect(definitionFor('electricity')?.compact?.({ zone: 'AT' })).toBe('electricity?zone=AT');
  });

  it('gives every compact view an example for the size previews (#596)', () => {
    for (const d of widgetDefinitions) {
      if (d.compact) expect(d.compactExample?.(), d.type.kind).toBeTruthy();
    }
  });
});
