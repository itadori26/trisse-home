import { describe, expect, it } from 'vitest';
import { groupByCategory, moveFocus, searchWidgetTypes } from './catalog';
import type { WidgetType } from './dashboard';

const type = (
  kind: string,
  title: string,
  category: WidgetType['category'],
  keywords: string[] = [],
  description = '',
): WidgetType => ({
  kind,
  title,
  description,
  category,
  keywords,
  defaultSize: { w: 4, h: 3 },
});

const types = [
  type('clock', 'Clock', 'everyday', ['time', 'Zeit']),
  type('weather', 'Weather', 'information', ['forecast'], 'Rain and sun for a place'),
  type('services', 'Services', 'server', ['uptime', 'status']),
  type('wireguard', 'WireGuard', 'network', ['vpn']),
  type('calendar', 'Calendar', 'information', ['events'], 'Termine aus CalDAV'),
];

describe('widget catalogue (#104)', () => {
  it('groups by category in a fixed order', () => {
    expect(groupByCategory(types).map((g) => [g.category, g.types.map((t) => t.kind)])).toEqual([
      ['everyday', ['clock']],
      ['information', ['weather', 'calendar']],
      ['server', ['services']],
      ['network', ['wireguard']],
    ]);
    expect(groupByCategory(types.slice(0, 1)).length).toBe(1);
  });

  it('searches title, keywords and description, ranked in that order', () => {
    expect(searchWidgetTypes(types, '').length).toBe(5);
    expect(searchWidgetTypes(types, 'vpn').map((t) => t.kind)).toEqual(['wireguard']);
    expect(searchWidgetTypes(types, 'ZEIT').map((t) => t.kind)).toEqual(['clock']);
    expect(searchWidgetTypes(types, 'rain').map((t) => t.kind)).toEqual(['weather']);
    expect(searchWidgetTypes(types, 'termine caldav').map((t) => t.kind)).toEqual(['calendar']);
    expect(searchWidgetTypes(types, 'nothing')).toEqual([]);
    // "s" matches the title of services (prefix) before others containing an "s".
    expect(searchWidgetTypes(types, 'serv')[0]?.kind).toBe('services');
  });

  it('ignores accents', () => {
    const t = [type('a', 'Météo', 'information')];
    expect(searchWidgetTypes(t, 'meteo').length).toBe(1);
  });

  it('moves focus through columns', () => {
    const cols = [2, 3, 1];
    expect(moveFocus(cols, [0, 1], 'ArrowDown')).toEqual([1, 0]);
    expect(moveFocus(cols, [1, 0], 'ArrowUp')).toEqual([0, 1]);
    expect(moveFocus(cols, [1, 2], 'ArrowRight')).toEqual([2, 0]);
    expect(moveFocus(cols, [2, 0], 'ArrowLeft')).toEqual([1, 0]);
    expect(moveFocus(cols, [2, 0], 'ArrowDown')).toEqual([2, 0]);
    expect(moveFocus(cols, [0, 0], 'ArrowUp')).toEqual([0, 0]);
    expect(moveFocus(cols, [1, 1], 'End')).toEqual([2, 0]);
    expect(moveFocus(cols, [1, 1], 'Home')).toEqual([0, 0]);
  });
});
