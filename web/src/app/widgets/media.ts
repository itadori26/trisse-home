import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The media widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'media',
      category: 'network',
      keywords: ['jellyfin', 'plex', 'movies', 'series', 'filme', 'serien'],
      get title() {
        return t('widget.media.title');
      },
      get description() {
        return t('widget.media.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'activity',
    view: lazy(
      () => import('../../features/media/ui/MediaWidget.svelte'),
      (config, c) => ({ config, media: c.media }),
    ),
    settings: lazy(
      () => import('../../features/media/ui/MediaWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
