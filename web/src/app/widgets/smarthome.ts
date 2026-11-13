import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The smarthome widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'smarthome',
      category: 'network',
      keywords: ['home assistant', 'switch', 'sensor', 'iot', 'smart home', 'hausautomation'],
      get title() {
        return t('widget.smarthome.title');
      },
      get description() {
        return t('widget.smarthome.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'activity',
    view: lazy(
      () => import('../../features/smarthome/ui/SmartHomeWidget.svelte'),
      (config, c) => ({ config, smarthome: c.smarthome }),
    ),
    settings: lazy(
      () => import('../../features/smarthome/ui/SmartHomeWidgetSettings.svelte'),
      (config, c) => ({ config, smarthome: c.smarthome }),
    ),
  },
];
