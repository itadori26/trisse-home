import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The releases widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'releases',
      category: 'information',
      keywords: ['music', 'musik', 'albums', 'musicbrainz', 'artists'],
      get title() {
        return t('widget.releases.title');
      },
      get description() {
        return t('widget.releases.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'activity',
    view: lazy(
      () => import('../../features/releases/ui/ReleasesWidget.svelte'),
      (config, c) => ({ config, releases: c.releases }),
    ),
    settings: lazy(
      () => import('../../features/releases/ui/ReleasesWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
