import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The FRITZ!Box call list, separate from the internet widget (#168). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'fritzbox-calls',
      category: 'network',
      keywords: ['avm', 'anrufe', 'calls', 'telefon', 'phone', 'anrufliste'],
      get title() {
        return t('widget.fritzbox_calls.title');
      },
      get description() {
        return t('widget.fritzbox_calls.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'history',
    view: lazy(
      () => import('../../features/fritzbox/ui/FritzBoxCallsWidget.svelte'),
      (config, c) => ({ config, fritzbox: c.fritzbox }),
    ),
    settings: lazy(
      () => import('../../features/fritzbox/ui/FritzBoxCallsSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
