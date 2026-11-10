import { describe, expect, it } from 'vitest';
import { formatDue, parseTodosConfig } from './todo';

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
