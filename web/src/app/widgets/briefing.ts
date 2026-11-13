import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The briefing widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'briefing',
      category: 'information',
      keywords: ['today', 'summary', 'morning', 'heute', 'tagesüberblick', 'zusammenfassung'],
      get title() {
        return t('widget.briefing.title');
      },
      get description() {
        return t('widget.briefing.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'sun',
    view: lazy(
      () => import('../../features/briefing/ui/BriefingWidget.svelte'),
      (_config, c) => ({ briefing: c.briefing }),
    ),
  },
];
