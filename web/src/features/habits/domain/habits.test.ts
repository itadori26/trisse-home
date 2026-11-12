import { describe, expect, it } from 'vitest';
import { editableDays, heatWeeks, isDue, isoDay, parseHabitsConfig, validateHabit } from './habits';

describe('habits domain', () => {
  it('computes today and yesterday in local time, across month ends', () => {
    expect(editableDays(new Date(2026, 11, 1, 0, 15))).toEqual({
      today: '2026-12-01',
      yesterday: '2026-11-30',
    });
    expect(isoDay(new Date(2026, 0, 5))).toBe('2026-01-05');
  });

  it('knows on which weekdays a habit is due', () => {
    expect(isDue({ days: [] }, 3)).toBe(true);
    expect(isDue({ days: [1, 3, 5] }, 3)).toBe(true);
    expect(isDue({ days: [1, 3, 5] }, 4)).toBe(false);
  });

  it('validates the name', () => {
    expect(validateHabit({ name: ' ', emoji: '', days: [] })).toBe('habits.invalid_name');
    expect(validateHabit({ name: 'x'.repeat(61), emoji: '', days: [] })).toBe('habits.invalid_name');
    expect(validateHabit({ name: 'Read', emoji: '📖', days: [] })).toBeNull();
  });

  it('splits the heatmap into weeks and reads the config', () => {
    expect(heatWeeks(Array.from({ length: 28 }, (_, i) => i)).map((w) => w.length)).toEqual([7, 7, 7, 7]);
    expect(parseHabitsConfig({})).toEqual({ showHeatmap: true });
    expect(parseHabitsConfig({ showHeatmap: false })).toEqual({ showHeatmap: false });
  });
});
