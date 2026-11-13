import { describe, expect, it } from 'vitest';
import { dayPart, greetingKey } from './greeting';

const at = (h: number, m = 0) => new Date(2026, 10, 12, h, m);

describe('dayPart', () => {
  it('splits the day at 5, 11, 18 and 23 o’clock', () => {
    expect(dayPart(at(4, 59))).toBe('night');
    expect(dayPart(at(5))).toBe('morning');
    expect(dayPart(at(10, 59))).toBe('morning');
    expect(dayPart(at(11))).toBe('day');
    expect(dayPart(at(18))).toBe('evening');
    expect(dayPart(at(23))).toBe('night');
    expect(dayPart(at(0))).toBe('night');
  });
});

describe('greetingKey', () => {
  it('uses the named variant only with a name', () => {
    expect(greetingKey('evening', true)).toBe('appearance.greeting_evening_named');
    expect(greetingKey('morning', false)).toBe('appearance.greeting_morning');
  });
});
