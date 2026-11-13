import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The generic JSON widget (#48). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'apiwidget',
      category: 'server',
      keywords: ['json', 'api', 'jsonpath', 'rest', 'http', 'endpoint', 'value', 'wert', 'schnittstelle'],
      get title() {
        return t('widget.apiwidget.title');
      },
      get description() {
        return t('widget.apiwidget.description');
      },
      defaultSize: { w: 3, h: 3 },
    },
    icon: 'gauge',
    view: lazy(
      () => import('../../features/apiwidget/ui/ApiWidget.svelte'),
      (config, c) => ({ config, apiwidget: c.apiwidget }),
    ),
    settings: lazy(
      () => import('../../features/apiwidget/ui/ApiWidgetConfig.svelte'),
      (config, c) => ({ config, apiwidget: c.apiwidget }),
    ),
  },
];
