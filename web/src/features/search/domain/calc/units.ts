import { evaluateTokens, tokenizeExpression } from './expression';

/**
 * Unit conversion for the smart search (#160): "5 km in mi",
 * "30 °C in F", "2 GiB to MB". A table of units per dimension with the
 * factor to the dimension's base unit; temperatures have an offset.
 * Data sizes: KB/MB/GB/TB are decimal (SI), KiB/MiB/GiB/TiB binary.
 */

export type Dimension = 'length' | 'mass' | 'volume' | 'temperature' | 'data' | 'time' | 'speed';

export interface Unit {
  /** Symbol shown in the answer. */
  symbol: string;
  dimension: Dimension;
  /** Value in base units = value × factor + offset. */
  factor: number;
  offset?: number;
  /** Names and symbols the user may type (lower case, without "°"). */
  aliases: string[];
}

const u = (symbol: string, dimension: Dimension, factor: number, aliases: string[], offset = 0): Unit => ({
  symbol,
  dimension,
  factor,
  offset,
  aliases: [symbol.toLowerCase(), ...aliases],
});

export const UNITS: readonly Unit[] = [
  // Length, base metre.
  u('mm', 'length', 0.001, ['millimeter', 'millimetre', 'millimetern']),
  u('cm', 'length', 0.01, ['zentimeter', 'centimeter', 'centimetre', 'zentimetern']),
  u('m', 'length', 1, ['meter', 'metre', 'metern', 'meters', 'metres']),
  u('km', 'length', 1000, ['kilometer', 'kilometre', 'kilometern', 'kilometers', 'kilometres']),
  u('in', 'length', 0.0254, ['inch', 'inches', 'zoll', '"']),
  u('ft', 'length', 0.3048, ['foot', 'feet', 'fuss']),
  u('yd', 'length', 0.9144, ['yard', 'yards']),
  u('mi', 'length', 1609.344, ['mile', 'miles', 'meile', 'meilen']),
  u('NM', 'length', 1852, ['nmi', 'seemeile', 'seemeilen']),
  // Mass, base kilogram.
  u('mg', 'mass', 1e-6, ['milligramm', 'milligram']),
  u('g', 'mass', 0.001, ['gramm', 'gram', 'grams']),
  u('kg', 'mass', 1, ['kilogramm', 'kilogram', 'kilograms', 'kilo']),
  u('t', 'mass', 1000, ['tonne', 'tonnen', 'tonnes']),
  u('oz', 'mass', 0.028349523125, ['ounce', 'ounces', 'unze', 'unzen']),
  u('lb', 'mass', 0.45359237, ['lbs', 'pound', 'pounds']),
  // Volume, base litre.
  u('ml', 'volume', 0.001, ['milliliter', 'millilitre']),
  u('cl', 'volume', 0.01, ['zentiliter', 'centiliter']),
  u('dl', 'volume', 0.1, ['deziliter', 'deciliter']),
  u('l', 'volume', 1, ['liter', 'litre', 'litern', 'liters', 'litres']),
  u('m³', 'volume', 1000, ['m3', 'kubikmeter', 'cubicmeter']),
  u('gal', 'volume', 3.785411784, ['gallon', 'gallons', 'gallone', 'gallonen']),
  u('fl oz', 'volume', 0.0295735295625, ['floz']),
  u('cup', 'volume', 0.2365882365, ['cups']),
  // Temperature, base kelvin.
  u('°C', 'temperature', 1, ['c', 'celsius', 'grad'], 273.15),
  u('°F', 'temperature', 5 / 9, ['f', 'fahrenheit'], 273.15 - (32 * 5) / 9),
  u('K', 'temperature', 1, ['kelvin']),
  // Data, base byte.
  u('B', 'data', 1, ['byte', 'bytes']),
  u('KB', 'data', 1e3, ['kilobyte', 'kilobytes']),
  u('MB', 'data', 1e6, ['megabyte', 'megabytes']),
  u('GB', 'data', 1e9, ['gigabyte', 'gigabytes']),
  u('TB', 'data', 1e12, ['terabyte', 'terabytes']),
  u('KiB', 'data', 1024, ['kibibyte']),
  u('MiB', 'data', 1024 ** 2, ['mebibyte']),
  u('GiB', 'data', 1024 ** 3, ['gibibyte']),
  u('TiB', 'data', 1024 ** 4, ['tebibyte']),
  // Time, base second.
  u('ms', 'time', 0.001, ['millisekunde', 'millisekunden', 'millisecond', 'milliseconds']),
  u('s', 'time', 1, ['sek', 'sec', 'sekunde', 'sekunden', 'second', 'seconds']),
  u('min', 'time', 60, ['minute', 'minuten', 'minutes']),
  u('h', 'time', 3600, ['std', 'stunde', 'stunden', 'hour', 'hours']),
  u('d', 'time', 86_400, ['tag', 'tage', 'tagen', 'day', 'days']),
  u('wk', 'time', 604_800, ['woche', 'wochen', 'week', 'weeks']),
  u('yr', 'time', 31_557_600, ['jahr', 'jahre', 'jahren', 'year', 'years']),
  // Speed, base metre per second.
  u('m/s', 'speed', 1, ['mps']),
  u('km/h', 'speed', 1 / 3.6, ['kmh', 'kph', 'stundenkilometer']),
  u('mph', 'speed', 0.44704, []),
  u('kn', 'speed', 1852 / 3600, ['knoten', 'knot', 'knots']),
];

const BY_ALIAS = new Map<string, Unit>();
for (const unit of UNITS) for (const a of unit.aliases) if (!BY_ALIAS.has(a)) BY_ALIAS.set(a, unit);

/** The unit for a typed name ("km", "°C", "Stunden"), or null. */
export function findUnit(text: string): Unit | null {
  const key = text.trim().toLowerCase().replace(/^°\s*/, '').replace(/\s+/g, '');
  return BY_ALIAS.get(key) ?? BY_ALIAS.get(text.trim().toLowerCase()) ?? null;
}

/** Converts a value; null when the units measure different things. */
export function convert(value: number, from: Unit, to: Unit): number | null {
  if (from.dimension !== to.dimension) return null;
  const base = value * from.factor + (from.offset ?? 0);
  return (base - (to.offset ?? 0)) / to.factor;
}

export interface Conversion {
  value: number;
  from: Unit;
  result: number;
  to: Unit;
}

// The trailing space is a lookahead, so "5 in in cm" finds both words.
const CONNECTOR = /\s+(?:in|to|nach|zu|als|as)(?=\s)|\s*(?:->|→|=)/gi;

/**
 * Reads "<amount> <unit> in <unit>". The amount may be an expression
 * ("2*3 km in m"); the last connector word separates the target, so
 * "5 in in cm" works.
 */
export function parseConversion(input: string): Conversion | null {
  const text = input.trim();
  const connectors = [...text.matchAll(CONNECTOR)];
  const last = connectors.at(-1);
  if (!last || last.index === undefined) return null;
  const to = findUnit(text.slice(last.index + last[0].length).trim());
  if (!to) return null;
  const source = text.slice(0, last.index).trim();
  // The unit is the trailing run of letters and unit signs after the amount.
  const m = /^(.*?\d\)?)\s*([^\d\s()][^\d()]*)$/.exec(source);
  if (!m) return null;
  const from = findUnit(m[2] ?? '');
  if (!from) return null;
  const tokens = tokenizeExpression(m[1] ?? '');
  const value = tokens ? evaluateTokens(tokens) : null;
  if (value === null) return null;
  const result = convert(value, from, to);
  if (result === null || !Number.isFinite(result)) return null;
  return { value, from, result, to };
}
