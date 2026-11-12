/** Local calendar date as YYYY-MM-DD. */
export function isoDate(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
}

/** today + n days; computed at noon so DST changes never shift the date. */
export function addDays(today: Date, n: number): Date {
  return new Date(today.getFullYear(), today.getMonth(), today.getDate() + n, 12);
}

const weekdays: Record<string, number> = {
  sonntag: 0,
  montag: 1,
  dienstag: 2,
  mittwoch: 3,
  donnerstag: 4,
  freitag: 5,
  samstag: 6,
  sonnabend: 6,
  sunday: 0,
  monday: 1,
  tuesday: 2,
  wednesday: 3,
  thursday: 4,
  friday: 5,
  saturday: 6,
};

const relative: Record<string, number> = {
  heute: 0,
  today: 0,
  morgen: 1,
  tomorrow: 1,
  übermorgen: 2,
  uebermorgen: 2,
};

/** A recognised date and the text without it. */
export interface DueResult {
  text: string;
  due: string; // YYYY-MM-DD or ''
}

interface Rule {
  re: RegExp;
  date: (m: RegExpExecArray, today: Date) => Date | null;
}

/** Capture group i as text ('' when absent). */
const g = (m: RegExpExecArray, i: number): string => m[i] ?? '';

const validDate = (y: number, m: number, d: number): Date | null => {
  const date = new Date(y, m - 1, d, 12);
  return date.getMonth() === m - 1 && date.getDate() === d ? date : null;
};

const rules: Rule[] = [
  // @2026-12-01 (older to-do syntax) and plain ISO dates
  { re: /(?:^|\s)@?(\d{4})-(\d{2})-(\d{2})(?=\s|$)/u, date: (m) => validDate(+g(m, 1), +g(m, 2), +g(m, 3)) },
  // German 14.2. / 14.2.2027
  {
    re: /(?:^|\s)(?:am\s+)?(\d{1,2})\.(\d{1,2})\.(\d{4})?(?=\s|$)/iu,
    date: (m, today) => {
      const year = m[3] ? +m[3] : today.getFullYear();
      const d = validDate(year, +g(m, 2), +g(m, 1));
      // Without a year a past date means next year.
      if (d && !m[3] && isoDate(d) < isoDate(today)) return validDate(year + 1, +g(m, 2), +g(m, 1));
      return d;
    },
  },
  // in 3 Tagen / in 3 days / in 2 Wochen / in 2 weeks
  {
    re: /(?:^|\s)in\s+(\d{1,3})\s+(tagen?|days?|wochen?|weeks?)(?=\s|$)/iu,
    date: (m, today) => addDays(today, +g(m, 1) * (/^w/i.test(g(m, 2)) ? 7 : 1)),
  },
  // in einer Woche / in a week / nächste Woche / next week
  {
    re: /(?:^|\s)(?:in\s+(?:einer|a)\s+(?:woche|week)|(?:nächste|naechste|next)\s+(?:woche|week))(?=\s|$)/iu,
    date: (_m, today) => addDays(today, 7),
  },
  // heute, morgen, übermorgen, today, tomorrow
  {
    re: /(?:^|\s)(übermorgen|uebermorgen|heute|morgen|today|tomorrow)(?=\s|$)/iu,
    date: (m, today) => addDays(today, relative[g(m, 1).toLowerCase()] ?? 0),
  },
  // (am|on|next|nächsten) Freitag / friday: the next such day after today
  {
    re: /(?:^|\s)(?:(?:am|on|next|nächsten|naechsten)\s+)?(montag|dienstag|mittwoch|donnerstag|freitag|samstag|sonnabend|sonntag|monday|tuesday|wednesday|thursday|friday|saturday|sunday)(?=\s|$)/iu,
    date: (m, today) => {
      const target = weekdays[g(m, 1).toLowerCase()] ?? 0;
      const delta = (target - today.getDay() + 7) % 7 || 7;
      return addDays(today, delta);
    },
  },
];

/**
 * Finds the first date phrase in German or English (#67, #105) and returns
 * the text without it: "Müll rausbringen morgen" → ("Müll rausbringen",
 * tomorrow). `today` is injected so the rules are testable.
 */
export function extractDue(text: string, today: Date): DueResult {
  for (const r of rules) {
    const m = r.re.exec(text);
    if (!m) continue;
    const d = r.date(m, today);
    if (!d) continue;
    const rest = (text.slice(0, m.index) + ' ' + text.slice(m.index + m[0].length))
      .replace(/\s+/g, ' ')
      .trim();
    return { text: rest, due: isoDate(d) };
  }
  return { text: text.trim(), due: '' };
}
