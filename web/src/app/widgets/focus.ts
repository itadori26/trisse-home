import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The focus widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'focus',
      category: 'everyday',
      keywords: ['pomodoro', 'timer', 'study', 'concentration', 'lernen', 'konzentration', 'pause'],
      get title() {
        return t('widget.focus.title');
      },
      get description() {
        return t('widget.focus.description');
      },
      defaultSize: { w: 3, h: 4 },
    },
    icon: 'timer',
    view: lazy(
      () => import('../../features/focus/ui/FocusWidget.svelte'),
      (config, c) => ({ config, focus: c.focus }),
    ),
    settings: lazy(
      () => import('../../features/focus/ui/FocusSettings.svelte'),
      (config, c) => ({ config, focus: c.focus }),
    ),
  },
];
