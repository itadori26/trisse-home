import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The holidays widget (#55). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'holidays',
      category: 'information',
      keywords: ['feiertage', 'ferien', 'schulferien', 'public holidays', 'school', 'vacation', 'bundesland'],
      get title() {
        return t('widget.holidays.title');
      },
      get description() {
        return t('widget.holidays.description');
      },
      defaultSize: { w: 3, h: 4 },
    },
    icon: 'calendar',
    view: lazy(
      () => import('../../features/holidays/ui/HolidaysWidget.svelte'),
      (config, c) => ({ config, holidays: c.holidays }),
    ),
    settings: lazy(
      () => import('../../features/holidays/ui/HolidaysWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
