import { ago, type BackupSource, type SourceState } from '../../features/backupmonitor/domain/source';
import type { State } from '../../features/search/domain/smart/answer';
import { t } from '../../shared/i18n';
import { defineSearch, MAX_ITEMS, terms } from '../searchKit';

const backupState: Record<SourceState, State> = {
  ok: 'ok',
  failed: 'error',
  overdue: 'warning',
  running: 'off',
  waiting: 'off',
};

/** When the job last succeeded, or when it failed. */
const last = (s: BackupSource, now: Date) =>
  s.state === 'failed'
    ? t('backupmonitor.failed_ago', { time: ago(s.lastFailureAt, now) })
    : ago(s.lastSuccessAt, now);

/**
 * Backups and heartbeats (#161): "backup", "ist das backup gelaufen",
 * "heartbeats" for the overview, a job name ("restic") for its status.
 */
export const search = defineSearch({
  module: 'backupmonitor',
  intents: [
    {
      id: 'backup.overview',
      source: 'backupmonitor',
      answer: 'list',
      widgetKind: 'backupmonitor',
      terms: [
        ...terms(
          3,
          'backup',
          'backups',
          'sicherung',
          'sicherungen',
          'heartbeat',
          'heartbeats',
          'cronjob',
          'cronjobs',
        ),
        ...terms(1.5, 'jobs', 'checks'),
        ...terms(1, 'gelaufen', 'letzte', 'status', 'failed', 'fehlgeschlagen'),
      ],
    },
    {
      id: 'backup.status',
      source: 'backupmonitor',
      prefixes: ['backup'],
      answer: 'status',
      widgetKind: 'backupmonitor',
      slots: [{ name: 'job', type: 'entity', entityKind: 'job', required: true }],
      terms: terms(1, 'backup', 'gelaufen', 'letzte', 'last', 'status'),
    },
  ],
  source: (c) => ({
    id: 'backupmonitor',
    async entities() {
      return (await c.backupMonitor.list()).map((s) => ({ kind: 'job', id: s.id, name: s.name }));
    },
    async resolve(m) {
      const list = await c.backupMonitor.list();
      const now = new Date();
      const base = { intent: m.intent.id, widgetKind: 'backupmonitor', at: now };
      if (m.intent.id === 'backup.overview') {
        const ok = list.filter((s) => s.state === 'ok').length;
        return {
          ...base,
          kind: 'list',
          title: t('smart.backups', { ok, total: list.length }),
          empty: t('backupmonitor.no_backups_monitored_yet_add_them'),
          // Problems first: the use case sorts failed and overdue to the top.
          items: list.slice(0, MAX_ITEMS).map((s) => ({
            label: s.name,
            value: `${t(`backupmonitor.state.${s.state}`)} · ${last(s, now)}`,
            state: backupState[s.state],
          })),
        };
      }
      const target = m.slots['job']?.type === 'entity' ? m.slots['job'].entity : null;
      const s = list.find((x) => x.id === target?.id);
      if (!s)
        return { ...base, at: null, kind: 'text', title: target?.name ?? '', text: t('smart.job_gone') };
      return {
        ...base,
        kind: 'status',
        title: s.name,
        state: backupState[s.state],
        text: t(`backupmonitor.state.${s.state}`),
        detail: [last(s, now), s.lastMessage].filter(Boolean).join(' · '),
      };
    },
  }),
});
