import { parcelName, type ParcelStatus, statusInfo } from '../../features/parcels/domain/parcel';
import type { State } from '../../features/search/domain/smart/answer';
import { intlLocale, t } from '../../shared/i18n';
import { defineSearch, MAX_ITEMS, terms } from '../searchKit';

const parcelState: Record<ParcelStatus, State | undefined> = {
  unknown: 'off',
  'pre-transit': undefined,
  transit: undefined,
  delivered: 'ok',
  failure: 'error',
};

/** Status of tracked parcels (#161): "paket", "wo ist mein paket", "dhl". */
export const search = defineSearch({
  module: 'parcels',
  intents: [
    {
      id: 'parcels.status',
      source: 'parcels',
      prefixes: ['paket', 'parcel'],
      answer: 'list',
      widgetKind: 'parcels',
      terms: [
        ...terms(3, 'paket', 'pakete', 'sendung', 'sendungen', 'parcel', 'parcels', 'package', 'packages'),
        ...terms(2, 'lieferung', 'delivery', 'dhl', 'tracking'),
        ...terms(1, 'kommt', 'coming', 'zugestellt', 'delivered'),
      ],
    },
  ],
  source: (c) => ({
    id: 'parcels',
    cacheMs: 60_000,
    async resolve(m) {
      const list = await c.parcels.list();
      const latest = list.reduce<Date | null>(
        (d, p) => (p.statusAt && (!d || p.statusAt > d) ? p.statusAt : d),
        null,
      );
      return {
        kind: 'list',
        intent: m.intent.id,
        widgetKind: 'parcels',
        title: t('smart.parcels'),
        at: latest,
        empty: t('smart.no_parcels'),
        items: list.slice(0, MAX_ITEMS).map((p) => ({
          label: parcelName(p),
          value: [
            p.error || t(statusInfo[p.status].label),
            p.estimatedDelivery && p.status !== 'delivered'
              ? t('smart.parcel_eta', { date: p.estimatedDelivery.toLocaleDateString(intlLocale()) })
              : p.location,
          ]
            .filter(Boolean)
            .join(' · '),
          state: p.error ? 'error' : parcelState[p.status],
        })),
      };
    },
  }),
});
