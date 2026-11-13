import { describe, expect, it } from 'vitest';
import { calculate } from './calc';
import { evaluate, parseNumber } from './expression';
import { convert, findUnit, parseConversion } from './units';

describe('parseNumber', () => {
  it.each([
    ['42', 42],
    ['1,5', 1.5],
    ['1.5', 1.5],
    ['1.000', 1],
    ['1.000.000', 1_000_000],
    ['1,000,000', 1_000_000],
    ['1.234,5', 1234.5],
    ['1,234.5', 1234.5],
    ['0,5', 0.5],
    ['0', 0],
  ])('%s → %d', (text, want) => expect(parseNumber(text)).toBe(want));

  it.each(['13.11.2026', '1.23.456', '1,2,3', '0151', '1.', '1,2.3,4'])('%s is no number', (text) =>
    expect(parseNumber(text)).toBeNull(),
  );
});

describe('evaluate', () => {
  it.each([
    ['2+3*4', 14],
    ['(2+3)*4', 20],
    ['2^3^2', 512],
    ['-2^2', -4],
    ['(-2)^2', 4],
    ['10-4-3', 3],
    ['100/4/5', 5],
    ['2 * -3', -6],
    ['1,5 + 2,25', 3.75],
    ['3 x 4', 12],
    ['3×4', 12],
    ['12 ÷ 4', 3],
    ['12 : 4', 3],
    ['19 % von 240', 45.6],
    ['19% of 240', 45.6],
    ['200 + 10 %', 220],
    ['200 - 10%', 180],
    ['200 * 10 %', 20],
    ['1/3*3', 1],
  ])('%s = %d', (input, want) => expect(evaluate(input)).toBeCloseTo(want, 10));

  it.each([
    ['1/0', 'division by zero'],
    ['5/(3-3)', 'division by zero in parentheses'],
    ['42', 'a bare number'],
    ['-5', 'a negative number'],
    ['50 %', 'a bare percentage'],
    ['2+', 'an incomplete expression'],
    ['(2+3', 'an open parenthesis'],
    ['2 3', 'two numbers without operator'],
    ['2026-11-13', 'an ISO date'],
    ['11/13/2026', 'a US date'],
    ['13.11.2026', 'a German date'],
    ['0151 1234 5678', 'a phone number'],
    ['0151-12345', 'a phone number with leading zero'],
    ['+49 151 12345', 'an international phone number'],
    ['2+abc', 'words'],
    ['10^400', 'an infinite result'],
    ['', 'nothing'],
  ])('%s gives null (%s)', (input) => expect(evaluate(input)).toBeNull());
});

describe('units', () => {
  it('finds units by symbol and name', () => {
    expect(findUnit('KM')?.symbol).toBe('km');
    expect(findUnit('°C')?.symbol).toBe('°C');
    expect(findUnit('° F')?.symbol).toBe('°F');
    expect(findUnit('Stunden')?.symbol).toBe('h');
    expect(findUnit('fl oz')?.symbol).toBe('fl oz');
    expect(findUnit('parsec')).toBeNull();
  });

  it('refuses to convert between dimensions', () => {
    const km = findUnit('km');
    const kg = findUnit('kg');
    expect(km && kg && convert(1, km, kg)).toBeNull();
  });

  it.each([
    ['5 km in mi', 3.10686],
    ['1 mi to km', 1.609344],
    ['30 °C in F', 86],
    ['30°C in °F', 86],
    ['-40 c in f', -40],
    ['212 F nach C', 100],
    ['0 K in °C', -273.15],
    ['2 GiB in MB', 2147.483648],
    ['1 GB in MiB', 953.6743],
    ['90 min in h', 1.5],
    ['1 Tag in Stunden', 24],
    ['100 km/h in mph', 62.1371],
    ['10 m/s → km/h', 36],
    ['1 l in ml', 1000],
    ['1 gallon in l', 3.785411784],
    ['1 lb in g', 453.59237],
    ['5 in in cm', 12.7],
    ['2*3 km in m', 6000],
    ['1,5 kg in g', 1500],
  ])('%s = %d', (input, want) => expect(parseConversion(input)?.result).toBeCloseTo(want, 3));

  it.each(['5 km in kg', '5 km', 'km in mi', '5 lightyears in km', 'restart jellyfin in 5'])(
    '%s gives null',
    (input) => expect(parseConversion(input)).toBeNull(),
  );
});

describe('calculate', () => {
  it('prefers a conversion and accepts "="', () => {
    expect(calculate('5 km in m')).toMatchObject({ kind: 'conversion', result: 5000 });
    expect(calculate('=2+2')).toMatchObject({ kind: 'expression', value: 4 });
    expect(calculate('2+2=')).toMatchObject({ kind: 'expression', value: 4 });
  });

  it('stays silent for normal searches', () => {
    for (const q of ['wetter berlin morgen', 'cpu', 'restart jellyfin', 'docker', 'iphone 15 pro', 'ps5'])
      expect(calculate(q)).toBeNull();
  });
});
