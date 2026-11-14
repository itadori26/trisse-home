import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The drop zone widget (#101). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'drop',
      category: 'everyday',
      keywords: [
        'share',
        'clipboard',
        'upload',
        'file',
        'datei',
        'teilen',
        'ablage',
        'airdrop',
        'zwischenablage',
      ],
      get title() {
        return t('widget.drop.title');
      },
      get description() {
        return t('widget.drop.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'package',
    view: lazy(
      () => import('../../features/drop/ui/DropWidget.svelte'),
      (_config, c) => ({ drop: c.drop }),
    ),
  },
];
