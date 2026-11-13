import { describe, expect, it } from 'vitest';
import { changedAgo, filterEntities, MAX_ENTITIES, parseWidgetConfig, stateText } from './smarthome';

describe('smart home domain', () => {
  it('reads only string entity IDs and caps the list', () => {
    expect(parseWidgetConfig({})).toEqual({ entities: [] });
    expect(parseWidgetConfig({ entities: ['switch.a', 3, null] })).toEqual({ entities: ['switch.a'] });
    const many = Array.from({ length: MAX_ENTITIES + 5 }, (_, i) => `sensor.s${i}`);
    expect(parseWidgetConfig({ entities: many }).entities).toHaveLength(MAX_ENTITIES);
  });
  it('shows values with their unit', () => {
    expect(stateText({ state: '42.5', unit: 'W' })).toBe('42.5 W');
    expect(stateText({ state: 'home', unit: '' })).toBe('home');
    expect(stateText({ state: 'unavailable', unit: 'W' })).not.toContain('W');
  });
  it('formats the last change relative to now', () => {
    const now = new Date('2026-11-12T12:00:00Z');
    expect(changedAgo(null, now)).toBe('');
    expect(changedAgo(new Date('2026-11-12T11:55:00Z'), now)).toContain('5');
  });
  it('filters by ID and name', () => {
    const list = [
      { id: 'switch.coffee', name: 'Kaffeemaschine' },
      { id: 'sensor.power', name: 'Power' },
    ];
    expect(filterEntities(list, 'kaffee').map((e) => e.id)).toEqual(['switch.coffee']);
    expect(filterEntities(list, 'SENSOR').map((e) => e.id)).toEqual(['sensor.power']);
    expect(filterEntities(list, ' ')).toHaveLength(2);
  });
});
