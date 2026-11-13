import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The fuel widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'fuel',
      category: 'everyday',
      keywords: ['petrol', 'gas', 'diesel', 'benzin', 'tanken', 'sprit', 'tankerkönig'],
      get title() {
        return t('widget.fuel.title');
      },
      get description() {
        return t('widget.fuel.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'gauge',
    view: lazy(
      () => import('../../features/fuel/ui/FuelWidget.svelte'),
      (config, c) => ({ config, fuel: c.fuel }),
    ),
    settings: lazy(
      () => import('../../features/fuel/ui/FuelWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
