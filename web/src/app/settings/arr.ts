import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'arr',
  label: 'settings.section.arr',
  module: 'arr',
  load: () => import('../../features/arr').then((m) => m.ArrSettings),
  props: (c) => ({ arr: c.arr }),
});
