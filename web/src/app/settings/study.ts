import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'study',
  label: 'settings.section.study',
  module: 'study',
  load: () => import('../../features/study').then((m) => m.StudySettings),
  props: (c) => ({ study: c.study }),
});
