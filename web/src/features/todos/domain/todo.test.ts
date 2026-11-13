import { describe, expect, it } from 'vitest';
import {
  describeReminder,
  describeRepeat,
  formatDue,
  NO_REPEAT,
  parseTodosConfig,
  validateSchedule,
} from './todo';

describe('formatDue', () => {
  const today = new Date(2026, 10, 10, 15, 0);
  it.each([
    ['', ''],
    ['2026-11-10', 'Today'],
    ['2026-11-11', 'Tomorrow'],
    ['2026-11-09', 'Yesterday'],
  ])('%s -> %s', (due, want) => expect(formatDue(due, today, 'en-US')).toBe(want));
  it('uses weekdays within a week', () => {
    expect(formatDue('2026-11-13', today, 'en-US')).toBe('Friday');
  });
  it('parses the config', () => {
    expect(parseTodosConfig({ showDone: true }).showDone).toBe(true);
  });
});

describe('recurrence and reminders (#68)', () => {
  const day = (d: number) => ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][d] ?? '';
  it('describes rules', () => {
    expect(describeRepeat({ kind: 'daily', every: 1, days: [] }, day)).toBe('Every day');
    expect(describeRepeat({ kind: 'weekly', every: 2, days: [4, 1] }, day)).toBe('Every 2 weeks on Mon, Thu');
    expect(describeRepeat({ kind: 'after', every: 10, days: [] }, day)).toBe('10 days after completion');
    expect(describeRepeat(NO_REPEAT, day)).toBe('');
    expect(describeReminder({ at: '08:00', daysBefore: 1 })).toBe('08:00, 1 day(s) before');
    expect(describeReminder({ at: '', daysBefore: 0 })).toBe('');
  });
  it.each([
    ['', { kind: 'daily', every: 0, days: [] }, { at: '', daysBefore: 0 }, /365/],
    ['', { kind: 'weekly', every: 1, days: [] }, { at: '', daysBefore: 0 }, /weekday/],
    ['', NO_REPEAT, { at: '08:00', daysBefore: 0 }, /due date/],
    ['2026-11-12', NO_REPEAT, { at: '8:00', daysBefore: 0 }, /08:00/],
    ['2026-11-12', NO_REPEAT, { at: '08:00', daysBefore: 31 }, /30/],
  ] as const)('rejects %j %j %j', (due, repeat, reminder, msg) =>
    expect(validateSchedule(due, { ...repeat, days: [...repeat.days] }, reminder)).toMatch(msg),
  );
  it('accepts valid schedules', () => {
    expect(
      validateSchedule('2026-11-12', { kind: 'monthly', every: 3, days: [] }, { at: '07:30', daysBefore: 2 }),
    ).toBeNull();
  });
});
