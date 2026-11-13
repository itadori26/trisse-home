import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The study progress widget (#57). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'study',
      category: 'everyday',
      keywords: ['ects', 'studium', 'uni', 'university', 'grades', 'noten', 'credits', 'module', 'semester'],
      get title() {
        return t('widget.study.title');
      },
      get description() {
        return t('widget.study.description');
      },
      defaultSize: { w: 3, h: 4 },
    },
    icon: 'activity',
    view: lazy(
      () => import('../../features/study/ui/StudyWidget.svelte'),
      (_config, c) => ({ study: c.study }),
    ),
  },
];
