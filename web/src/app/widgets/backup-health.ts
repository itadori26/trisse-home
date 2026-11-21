import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The traffic light of Waylight's own backups (#677); only the owner can read it. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'backup-health',
      category: 'server',
      keywords: ['backup', 'restore', 'sicherung', 'ampel', 'traffic light'],
      get title() {
        return t('widget.backup-health.title');
      },
      get description() {
        return t('widget.backup-health.description');
      },
      defaultSize: { w: 3, h: 2 },
    },
    icon: 'archive',
    compact: () => 'backup',
    compactExample: () => t('compact.example.backup-health'),
    alerts: true,
    view: lazy(
      () => import('../../features/backup/ui/BackupHealthWidget.svelte'),
      (_config, c) => ({ backups: c.backup }),
    ),
  },
];
