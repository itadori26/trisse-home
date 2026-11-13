import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The countdown widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'countdown',
      category: 'everyday',
      keywords: ['days', 'deadline', 'exam', 'holiday', 'countdown', 'tage', 'frist', 'prüfung', 'urlaub'],
      get title() {
        return t('widget.countdown.title');
      },
      get description() {
        return t('widget.countdown.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'hourglass',
    view: lazy(
      () => import('../../features/countdown/ui/CountdownWidget.svelte'),
      (config, c) => ({ config, countdown: c.countdown }),
    ),
    settings: lazy(
      () => import('../../features/countdown/ui/CountdownSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
