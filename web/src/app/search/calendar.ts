import type { CalendarEvent } from '../../features/calendar/domain/calendar';
import { intlLocale, t } from '../../shared/i18n';
import { defineSearch, MAX_ITEMS, terms } from '../searchKit';

const time = (d: Date) => d.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' });

/** Title of an agenda day: today, tomorrow or the weekday. */
function dayTitle(offset: number, date: Date): string {
  if (offset === 0) return t('time.today');
  if (offset === 1) return t('time.tomorrow');
  return date.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'long' });
}

/** Events that touch the given local day, all-day events first. */
export function eventsOnDay(events: readonly CalendarEvent[], day: Date): CalendarEvent[] {
  const start = new Date(day.getFullYear(), day.getMonth(), day.getDate());
  const end = new Date(start.getFullYear(), start.getMonth(), start.getDate() + 1);
  return events
    .filter((e) => e.start < end && e.end > start)
    .sort((a, b) => Number(b.allDay) - Number(a.allDay) || a.start.getTime() - b.start.getTime());
}

/** Agenda of a day (#161): "termine", "heute", "morgen", "termine am freitag". */
export const search = defineSearch({
  module: 'calendar',
  intents: [
    {
      id: 'calendar.agenda',
      source: 'calendar',
      prefixes: ['termine', 'kalender', 'calendar'],
      answer: 'list',
      widgetKind: 'calendar',
      // "morgen" alone is the agenda, "wetter morgen" is not.
      slots: [{ name: 'day', type: 'day', alone: 1.5 }],
      terms: [
        ...terms(3, 'termin', 'termine', 'kalender', 'calendar', 'agenda', 'appointments', 'meetings'),
        ...terms(1, 'vor', 'geplant', 'planned', 'schedule'),
      ],
    },
  ],
  source: (c) => ({
    id: 'calendar',
    cacheMs: 60_000,
    async resolve(m) {
      const offset = m.slots['day']?.type === 'day' ? m.slots['day'].offset : 0;
      const now = new Date();
      const day = new Date(now.getFullYear(), now.getMonth(), now.getDate() + offset);
      const agenda = await c.calendar.events(offset + 1);
      return {
        kind: 'list',
        intent: m.intent.id,
        widgetKind: 'calendar',
        title: `${t('smart.agenda')} · ${dayTitle(offset, day)}`,
        at: new Date(),
        empty: agenda.error || t('smart.no_events'),
        items: eventsOnDay(agenda.events, day)
          .slice(0, MAX_ITEMS)
          .map((e) => ({
            label: e.summary,
            value: [e.allDay ? t('calendar.all_day_short') : `${time(e.start)}–${time(e.end)}`, e.location]
              .filter(Boolean)
              .join(' · '),
          })),
      };
    },
  }),
});
