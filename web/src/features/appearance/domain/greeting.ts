/** Part of the day that decides the greeting (#167). */
export type DayPart = 'morning' | 'day' | 'evening' | 'night';

/** 5–11 morning, 11–18 day, 18–23 evening, otherwise night (local time). */
export function dayPart(now: Date): DayPart {
  const h = now.getHours();
  if (h >= 5 && h < 11) return 'morning';
  if (h >= 11 && h < 18) return 'day';
  if (h >= 18 && h < 23) return 'evening';
  return 'night';
}

/** Message key of the greeting; with or without a name. */
export function greetingKey(part: DayPart, named: boolean): `appearance.greeting_${DayPart}${'' | '_named'}` {
  return named ? `appearance.greeting_${part}_named` : `appearance.greeting_${part}`;
}
