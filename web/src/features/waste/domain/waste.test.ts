import { describe, expect, it } from 'vitest';
import { dueLabel, isDueSoon } from './waste';

describe('waste domain', () => {
  it('labels due dates', () => {
    expect(dueLabel(0)).toBe('Today');
    expect(dueLabel(1)).toBe('Tomorrow');
    expect(dueLabel(5)).toBe('in 5 days');
  });
  it('highlights tomorrow and this morning', () => {
    const evening = new Date(2026, 10, 11, 19);
    const morning = new Date(2026, 10, 11, 7);
    expect(isDueSoon({ daysUntil: 1 }, evening)).toBe(true);
    expect(isDueSoon({ daysUntil: 0 }, morning)).toBe(true);
    expect(isDueSoon({ daysUntil: 0 }, evening)).toBe(false);
    expect(isDueSoon({ daysUntil: 3 }, evening)).toBe(false);
  });
});
