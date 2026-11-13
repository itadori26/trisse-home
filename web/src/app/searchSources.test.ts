import { describe, expect, it } from 'vitest';
import { calcSource } from './searchSources';

describe('calcSource (#160)', () => {
  const direct = (q: string) => calcSource().direct?.(q) ?? null;

  it('answers arithmetic with a copyable plain value', () => {
    const a = direct('1000*1000+0.5');
    expect(a).toMatchObject({ kind: 'metric', intent: 'calc.expression', title: '1000*1000+0.5 =' });
    // No thousands separators in the clipboard value.
    expect(a?.copy).toMatch(/^1000000[.,]5$/);
    expect(direct('1/3')?.copy).toMatch(/^0[.,]3333333333$/);
  });

  it('rounds conversions to six significant digits', () => {
    const a = direct('5 km in mi');
    expect(a).toMatchObject({ intent: 'calc.conversion', value: expect.stringMatching(/^3[.,]10686 mi$/) });
    expect(a?.copy).toMatch(/^3[.,]10686$/);
    expect(direct('30 °C in F')).toMatchObject({ value: '86°F' });
  });

  it('stays silent for other queries', () => {
    expect(direct('wetter morgen')).toBeNull();
    expect(direct('42')).toBeNull();
  });
});
