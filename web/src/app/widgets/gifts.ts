import type { BirthdaySource } from '../../features/gifts/usecase/giftsUseCases';
import { t } from '../../shared/i18n';
import type { Container } from '../container';
import { lazy, type WidgetDefinition } from '../widgetKit';

/**
 * The next birthdays as the gift ideas' port: the features stay independent,
 * and without the birthdays module the list is empty.
 */
function birthdaySource(c: Container): BirthdaySource {
  return {
    async upcoming() {
      if (!(await c.enabledModules()).modules.has('birthdays')) return [];
      return (await c.birthdays.upcoming(0)).map((b) => ({ name: b.name, next: b.next, inDays: b.inDays }));
    },
  };
}

/** The gift ideas widget (#684). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'gifts',
      category: 'everyday',
      keywords: [
        'gift',
        'present',
        'birthday',
        'christmas',
        'family',
        'geschenk',
        'geburtstag',
        'weihnachten',
        'familie',
      ],
      get title() {
        return t('widget.gifts.title');
      },
      get description() {
        return t('widget.gifts.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'package',
    compact: () => 'gifts',
    compactExample: () => t('compact.example.gifts'),
    view: lazy(
      () => import('../../features/gifts/ui/GiftsWidget.svelte'),
      (_config, c) => ({ gifts: c.gifts, birthdays: birthdaySource(c) }),
    ),
  },
];
