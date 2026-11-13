import { evaluate } from './expression';
import { type Conversion, parseConversion } from './units';

/** Result of the calculator (#160): an arithmetic value or a unit conversion. */
export type CalcResult =
  { kind: 'expression'; expression: string; value: number } | ({ kind: 'conversion' } & Conversion);

/**
 * Calculates a search query: a unit conversion ("5 km in mi") or an
 * arithmetic expression ("2+3*4", "19 % von 240"). A leading or trailing
 * "=" is allowed. Everything else yields null, so the calculator stays
 * silent while the user types something different.
 */
export function calculate(query: string): CalcResult | null {
  const text = query.trim().replace(/^=\s*/, '').replace(/\s*=$/, '');
  if (!/\d/.test(text) || text.length > 200) return null;
  const conversion = parseConversion(text);
  if (conversion) return { kind: 'conversion', ...conversion };
  const value = evaluate(text);
  return value === null ? null : { kind: 'expression', expression: text, value };
}
