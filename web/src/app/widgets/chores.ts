import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The chore plan widget (#683). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'chores',
      category: 'everyday',
      keywords: [
        'chore',
        'household',
        'cleaning',
        'rota',
        'family',
        'haushalt',
        'putzplan',
        'aufgaben',
        'familie',
        'ämter',
      ],
      get title() {
        return t('widget.chores.title');
      },
      get description() {
        return t('widget.chores.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'done',
    compact: () => 'chores',
    compactExample: () => t('compact.example.chores'),
    view: lazy(
      () => import('../../features/chores/ui/ChoresWidget.svelte'),
      (_config, c) => ({ chores: c.chores }),
    ),
  },
];
