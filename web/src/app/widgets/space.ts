import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The space widget (#64). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'space',
      category: 'information',
      keywords: [
        'rocket',
        'launch',
        'nasa',
        'apod',
        'rakete',
        'raumfahrt',
        'weltraum',
        'astronomy',
        'spacex',
      ],
      get title() {
        return t('widget.space.title');
      },
      get description() {
        return t('widget.space.description');
      },
      defaultSize: { w: 4, h: 6 },
    },
    icon: 'sun',
    view: lazy(
      () => import('../../features/space/ui/SpaceWidget.svelte'),
      (config, c) => ({ config, space: c.space }),
    ),
    settings: lazy(
      () => import('../../features/space/ui/SpaceWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
