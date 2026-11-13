import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The habit tracker widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'habits',
      category: 'everyday',
      keywords: ['habit', 'streak', 'routine', 'gewohnheit', 'routine', 'serie', 'tracker'],
      get title() {
        return t('widget.habits.title');
      },
      get description() {
        return t('widget.habits.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'check',
    view: lazy(
      () => import('../../features/habits/ui/HabitsWidget.svelte'),
      (config, c) => ({ config, habits: c.habits }),
    ),
    settings: lazy(
      () => import('../../features/habits/ui/HabitsSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
