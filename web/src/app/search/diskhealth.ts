import type { DiskStatus } from '../../features/diskhealth/domain/disk';
import type { State } from '../../features/search/domain/smart/answer';
import { t } from '../../shared/i18n';
import { defineSearch, MAX_ITEMS, terms } from '../searchKit';

const diskState: Record<DiskStatus, State> = {
  ok: 'ok',
  warning: 'warning',
  critical: 'error',
  unknown: 'off',
};

/**
 * SMART health of the disks (#161): "smart", "festplatte kaputt". "Platte
 * voll" stays with the disk usage of the system module and its forecast.
 */
export const search = defineSearch({
  module: 'diskhealth',
  intents: [
    {
      id: 'diskhealth.status',
      source: 'diskhealth',
      prefixes: ['smart'],
      answer: 'list',
      widgetKind: 'diskhealth',
      terms: [
        ...terms(3, 'smart', 'diskhealth', 'festplattengesundheit', 'gesundheit', 'health'),
        ...terms(2.5, 'zustand', 'defekt', 'kaputt', 'broken', 'failing'),
        ...terms(1, 'platte', 'festplatte', 'disk', 'ssd', 'hdd', 'nvme'),
      ],
    },
  ],
  source: (c) => ({
    id: 'diskhealth',
    cacheMs: 300_000,
    async resolve(m) {
      const disks = await c.diskhealth.disks();
      const latest = disks.reduce<Date | null>((d, x) => (!d || x.checkedAt > d ? x.checkedAt : d), null);
      return {
        kind: 'list',
        intent: m.intent.id,
        widgetKind: 'diskhealth',
        title: t('smart.disk_health'),
        at: latest,
        empty: t('diskhealth.no_disks'),
        items: disks.slice(0, MAX_ITEMS).map((d) => ({
          label: d.model || d.device,
          value: [t(`diskhealth.status.${d.status}`), d.temperature !== null ? `${d.temperature} °C` : '']
            .filter(Boolean)
            .join(' · '),
          state: diskState[d.status],
        })),
      };
    },
  }),
});
