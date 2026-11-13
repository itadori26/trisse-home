import { describe, expect, it } from 'vitest';
import {
  type BriefingSettings,
  isSectionKind,
  itemDetail,
  itemTime,
  moveEntry,
  sectionTitle,
  validateSettings,
} from './briefing';

const settings = (s: Partial<BriefingSettings> = {}): BriefingSettings => ({
  layout: [],
  pushOn: false,
  pushAt: '07:00',
  pushDays: [],
  skipEmpty: true,
  lat: 52.5,
  lon: 13.4,
  available: [],
  ...s,
});

describe('briefing domain (#52)', () => {
  it('knows its sections', () => {
    expect(isSectionKind('todos')).toBe(true);
    expect(isSectionKind('nope')).toBe(false);
    expect(sectionTitle('incidents')).toBe('Since yesterday evening');
  });

  it('moves layout entries within bounds', () => {
    const layout = [
      { kind: 'todos' as const, enabled: true },
      { kind: 'waste' as const, enabled: false },
    ];
    expect(moveEntry(layout, 1, -1)?.map((e) => e.kind)).toEqual(['waste', 'todos']);
    expect(moveEntry(layout, 0, -1)).toBeNull();
    expect(moveEntry(layout, 1, 1)).toBeNull();
  });

  it('validates time and location', () => {
    expect(validateSettings(settings())).toBeNull();
    expect(validateSettings(settings({ pushAt: '7:00' }))).toMatch(/07:00/);
    expect(validateSettings(settings({ lat: 91 }))).toMatch(/Latitude/);
    expect(validateSettings(settings({ lon: Number.NaN }))).toMatch(/Latitude/);
  });

  it('shows clock times only for appointments and incidents', () => {
    const item = { text: 'x', detail: '', at: new Date(2026, 10, 12, 9, 30), severity: 'info' as const };
    expect(itemTime('events', item, 'de-DE')).toBe('09:30');
    expect(itemTime('todos', item, 'de-DE')).toBe('');
    expect(itemTime('events', { ...item, at: null })).toBe('');
  });
});

describe('itemDetail', () => {
  const item = (detail: string) => ({ text: 'x', detail, at: null, severity: 'info' as const });
  it('translates the markers of to-dos and collections', () => {
    expect(itemDetail('todos', item('overdue'))).toBe('overdue');
    expect(itemDetail('waste', item('tomorrow'))).toBe('tomorrow');
  });
  it('keeps free text and markers of other sections as they are', () => {
    expect(itemDetail('events', item('today'))).toBe('today');
    expect(itemDetail('services', item('connection refused'))).toBe('connection refused');
  });
});
