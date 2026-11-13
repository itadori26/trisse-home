import { ago } from '../../features/backupmonitor/domain/source';
import type { State } from '../../features/search/domain/smart/answer';
import type { EventSeverity } from '../../features/timeline/domain/timeline';
import { t } from '../../shared/i18n';
import { defineSearch, terms } from '../searchKit';

const eventState: Record<EventSeverity, State> = { info: 'off', warning: 'warning', critical: 'error' };

/** Latest events of the timeline (#161): "was ist passiert", "vorfälle". */
export const search = defineSearch({
  module: 'timeline',
  intents: [
    {
      id: 'timeline.recent',
      source: 'timeline',
      prefixes: ['timeline'],
      answer: 'list',
      widgetKind: 'timeline',
      terms: [
        ...terms(3, 'timeline', 'vorfalle', 'vorfall', 'ereignisse', 'incidents', 'events', 'verlauf'),
        ...terms(2, 'passiert', 'happened', 'los'),
        ...terms(1, 'zuletzt', 'recent', 'recently', 'neues'),
      ],
    },
  ],
  source: (c) => ({
    id: 'timeline',
    async resolve(m) {
      const events = await c.timeline.list({ source: '', severity: '', limit: 5 });
      const now = new Date();
      return {
        kind: 'list',
        intent: m.intent.id,
        widgetKind: 'timeline',
        title: t('smart.timeline'),
        at: events[0]?.at ?? null,
        empty: t('timeline.empty'),
        items: events.map((e) => ({
          label: e.subject,
          value: [ago(e.at, now), e.message].filter(Boolean).join(' · '),
          state: eventState[e.severity],
        })),
      };
    },
  }),
});
