import { describe, expect, it } from 'vitest';
import { isSoon, parseCountdownConfig, remainingAt, target, validateInput } from './countdown';

describe('remainingAt', () => {
  const now = new Date(2026, 11, 20, 10, 0); // 20 Dec, 10:00 local time

  it('counts calendar days for later days', () => {
    expect(remainingAt({ date: '2026-12-24', time: '' }, now)).toEqual({ kind: 'days', days: 4 });
    expect(remainingAt({ date: '2026-12-21', time: '08:00' }, now)).toEqual({
      kind: 'hours',
      hours: 22,
      minutes: 0,
    });
    expect(remainingAt({ date: '2026-12-22', time: '09:00' }, now)).toEqual({ kind: 'days', days: 2 });
  });

  it('shows hours and minutes on the last day of a timed countdown', () => {
    expect(remainingAt({ date: '2026-12-20', time: '12:30' }, now)).toEqual({
      kind: 'hours',
      hours: 2,
      minutes: 30,
    });
    expect(remainingAt({ date: '2026-12-20', time: '10:00' }, now)).toEqual({ kind: 'past' });
  });

  it('treats a whole-day countdown as running on its day', () => {
    expect(remainingAt({ date: '2026-12-20', time: '' }, now)).toEqual({ kind: 'today' });
    expect(remainingAt({ date: '2026-12-19', time: '' }, now)).toEqual({ kind: 'past' });
  });

  it('builds the target in local time', () => {
    expect(target({ date: '2026-12-24', time: '18:15' })).toEqual(new Date(2026, 11, 24, 18, 15));
  });
});

describe('isSoon', () => {
  it('highlights fewer days than the threshold, today and the last hours', () => {
    expect(isSoon({ kind: 'days', days: 6 }, 7)).toBe(true);
    expect(isSoon({ kind: 'days', days: 7 }, 7)).toBe(false);
    expect(isSoon({ kind: 'today' }, 0)).toBe(true);
    expect(isSoon({ kind: 'hours', hours: 1, minutes: 0 }, 0)).toBe(true);
    expect(isSoon({ kind: 'past' }, 7)).toBe(false);
  });
});

describe('parseCountdownConfig', () => {
  it('falls back to defaults and clamps', () => {
    expect(parseCountdownConfig({})).toEqual({ highlightDays: 7, graceDays: 1 });
    expect(parseCountdownConfig({ highlightDays: 3, graceDays: 9999 })).toEqual({
      highlightDays: 3,
      graceDays: 365,
    });
    expect(parseCountdownConfig({ highlightDays: 'x', graceDays: -1 })).toEqual({
      highlightDays: 7,
      graceDays: 0,
    });
  });
});

describe('validateInput', () => {
  const ok = { title: 'Exam', emoji: '', date: '2026-12-24', time: '' };
  it('accepts valid input and rejects the rest', () => {
    expect(validateInput(ok)).toBeNull();
    expect(validateInput({ ...ok, title: ' ' })).toBe('countdown.invalid_title');
    expect(validateInput({ ...ok, date: '24.12.2026' })).toBe('countdown.invalid_date');
    expect(validateInput({ ...ok, time: '7:00' })).toBe('countdown.invalid_time');
    expect(validateInput({ ...ok, time: '07:00' })).toBeNull();
  });
});
