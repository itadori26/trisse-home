import { dueLabel } from '../../features/waste/domain/waste';
import { t } from '../../shared/i18n';
import { defineSearch, MAX_ITEMS, terms } from '../searchKit';

/** Next waste collections (#161): "müll", "gelbe tonne", "/abfall". */
export const search = defineSearch({
  module: 'waste',
  intents: [
    {
      id: 'waste.next',
      source: 'waste',
      prefixes: ['muell', 'mull', 'abfall', 'waste'],
      answer: 'list',
      widgetKind: 'waste',
      terms: [
        ...terms(3, 'mull', 'muell', 'abfall', 'mullabfuhr', 'tonne', 'waste', 'garbage', 'trash', 'bin'),
        ...terms(2, 'gelbe', 'gelber', 'sack', 'biotonne', 'restmull', 'papiertonne', 'recycling'),
        ...terms(1, 'abholung', 'abfuhr', 'collection', 'raus', 'out'),
      ],
    },
  ],
  source: (c) => ({
    id: 'waste',
    cacheMs: 300_000,
    async resolve(m) {
      const o = await c.waste.upcoming();
      return {
        kind: 'list',
        intent: m.intent.id,
        widgetKind: 'waste',
        title: t('smart.waste'),
        at: new Date(),
        empty: o.error || t('waste.no_upcoming_collections'),
        items: [...o.types]
          .sort((a, b) => a.daysUntil - b.daysUntil)
          .slice(0, MAX_ITEMS)
          .map((w) => ({
            label: `${w.icon} ${w.label}`.trim(),
            value: dueLabel(w.daysUntil),
            state: w.daysUntil <= 1 ? 'warning' : undefined,
          })),
      };
    },
  }),
});
