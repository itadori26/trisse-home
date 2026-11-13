import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'gamedeals',
  label: 'settings.section.gamedeals',
  module: 'gamedeals',
  load: () => import('../../features/gamedeals').then((m) => m.GameDealsSettings),
  props: (c) => ({ gamedeals: c.gamedeals }),
});
