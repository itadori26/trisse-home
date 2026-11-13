import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The read-later widget (#99). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'readlater',
      category: 'everyday',
      keywords: ['pocket', 'instapaper', 'article', 'artikel', 'lesen', 'read later', 'reader', 'bookmark'],
      get title() {
        return t('widget.readlater.title');
      },
      get description() {
        return t('widget.readlater.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'bookmark',
    view: lazy(
      () => import('../../features/readlater/ui/ReadLaterWidget.svelte'),
      (config, c) => ({ config, readlater: c.readlater }),
    ),
    settings: lazy(
      () => import('../../features/readlater/ui/ReadLaterWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
