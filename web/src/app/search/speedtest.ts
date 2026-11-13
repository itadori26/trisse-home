import { belowThreshold, formatMbps } from '../../features/speedtest/domain/speedtest';
import { t } from '../../shared/i18n';
import { defineSearch, terms } from '../searchKit';

/** Last speed test (#161): "speedtest", "wie schnell ist das internet". */
export const search = defineSearch({
  module: 'speedtest',
  intents: [
    {
      id: 'speedtest.latest',
      source: 'speedtest',
      prefixes: ['speed', 'speedtest'],
      answer: 'metric',
      widgetKind: 'speedtest',
      terms: [
        ...terms(3, 'speedtest', 'speed', 'bandbreite', 'bandwidth', 'geschwindigkeit'),
        ...terms(2, 'internet', 'download', 'upload', 'dsl', 'leitung', 'mbit'),
        ...terms(1, 'schnell', 'fast', 'langsam', 'slow'),
      ],
    },
  ],
  source: (c) => ({
    id: 'speedtest',
    cacheMs: 60_000,
    async resolve(m) {
      const base = { intent: m.intent.id, widgetKind: 'speedtest', title: t('smart.speedtest') };
      const [r, settings] = await Promise.all([
        c.speedtest.results(24),
        c.speedtest.settings().catch(() => null),
      ]);
      const l = r.latest;
      if (!l) return { ...base, at: null, kind: 'text', text: r.error || t('speedtest.no_results') };
      return {
        ...base,
        at: l.at,
        kind: 'metric',
        value: `↓ ${formatMbps(l.downloadMbps)} · ↑ ${formatMbps(l.uploadMbps)} Mbit/s`,
        detail: `${t('speedtest.ping')} ${Math.round(l.pingMs)} ms`,
        severity: settings && belowThreshold(l, settings) ? 'warning' : undefined,
      };
    },
  }),
});
