import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The arr widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'arr',
      category: 'server',
      keywords: ['sonarr', 'radarr', 'prowlarr', 'downloads', 'series', 'movies'],
      get title() {
        return t('widget.arr.title');
      },
      get description() {
        return t('widget.arr.description');
      },
      defaultSize: { w: 4, h: 6 },
    },
    icon: 'package',
    view: lazy(
      () => import('../../features/arr/ui/ArrWidget.svelte'),
      (config, c) => ({ config, arr: c.arr }),
    ),
    settings: lazy(
      () => import('../../features/arr/ui/ArrWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
