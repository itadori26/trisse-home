import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'holidays',
  label: 'settings.section.holidays',
  module: 'holidays',
  load: () => import('../../features/holidays').then((m) => m.HolidaysSettings),
  props: (c) => ({ holidays: c.holidays }),
});
