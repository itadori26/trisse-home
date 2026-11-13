import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The gamedeals widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'gamedeals',
      category: 'information',
      keywords: ['steam', 'epic', 'games', 'spiele', 'sale', 'free'],
      get title() {
        return t('widget.gamedeals.title');
      },
      get description() {
        return t('widget.gamedeals.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'cart',
    view: lazy(
      () => import('../../features/gamedeals/ui/GameDealsWidget.svelte'),
      (_config, c) => ({ gamedeals: c.gamedeals }),
    ),
  },
];
