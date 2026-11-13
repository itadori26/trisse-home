/** Part of the day that decides the greeting (#167). */
export type DayPart = 'morning' | 'afternoon' | 'evening' | 'night';

/** Morning 5–11, afternoon 12–17, evening 18–22, otherwise night (local hours, #189). */
export function dayPart(now: Date): DayPart {
  const h = now.getHours();
  if (h >= 5 && h <= 11) return 'morning';
  if (h >= 12 && h <= 17) return 'afternoon';
  if (h >= 18 && h <= 22) return 'evening';
  return 'night';
}

/** Message key of the greeting; with or without a name. */
export function greetingKey(part: DayPart, named: boolean): `appearance.greeting_${DayPart}${'' | '_named'}` {
  return named ? `appearance.greeting_${part}_named` : `appearance.greeting_${part}`;
}
