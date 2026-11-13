import { minutesUntil, parseTransitConfig } from '../../features/transit/domain/transit';
import { t } from '../../shared/i18n';
import { defineSearch, MAX_ITEMS, terms } from '../searchKit';

/**
 * Next departures (#161): "abfahrt", "nächster bus", "abfahrten hauptbahnhof".
 * A named stop is looked up; without one the stop of the first transit
 * widget is used.
 */
export const search = defineSearch({
  module: 'transit',
  intents: [
    {
      id: 'transit.departures',
      source: 'transit',
      prefixes: ['abfahrt', 'bus', 'departures'],
      answer: 'list',
      widgetKind: 'transit',
      slots: [{ name: 'place', type: 'place' }],
      terms: [
        ...terms(
          3,
          'abfahrt',
          'abfahrten',
          'departure',
          'departures',
          'bus',
          'tram',
          'strassenbahn',
          'ubahn',
          'sbahn',
        ),
        ...terms(2, 'bahn', 'zug', 'train', 'haltestelle', 'stop'),
        ...terms(1, 'nachste', 'next', 'fahrt'),
      ],
    },
  ],
  source: (c, ctx) => ({
    id: 'transit',
    cacheMs: 30_000,
    async resolve(m) {
      const base = { intent: m.intent.id, widgetKind: 'transit' };
      const named = m.slots['place']?.type === 'place' ? m.slots['place'].text : '';
      let stop: { id: string; name: string } | null = null;
      if (named) {
        const found = (await c.transit.searchStops(named))[0];
        if (!found)
          return { ...base, at: null, kind: 'text', title: named, text: t('transit.no_stops_found') };
        stop = found;
      } else {
        const cfg = parseTransitConfig(ctx.widgetConfig('transit') ?? {});
        if (cfg) stop = { id: cfg.stopId, name: cfg.stopName };
      }
      if (!stop)
        return {
          ...base,
          at: null,
          kind: 'text',
          title: t('transit.departures'),
          text: t('smart.which_stop'),
        };
      const now = new Date();
      const deps = await c.transit.departures({ stopId: stop.id, stopName: stop.name, count: MAX_ITEMS });
      return {
        ...base,
        at: now,
        kind: 'list',
        title: stop.name,
        empty: t('transit.no_departures_in_the_near_future'),
        items: deps.map((d) => ({
          label: `${d.line} → ${d.headsign}`,
          value: d.cancelled
            ? t('transit.cancelled')
            : [minutesUntil(d, now), d.delayMinutes > 0 ? `+${d.delayMinutes}` : '', d.track]
                .filter(Boolean)
                .join(' · '),
          state: d.cancelled ? 'error' : d.delayMinutes >= 5 ? 'warning' : undefined,
        })),
      };
    },
  }),
});
