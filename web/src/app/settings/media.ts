import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'media',
  label: 'settings.section.media',
  module: 'media',
  load: () => import('../../features/media').then((m) => m.MediaSettings),
  props: (c) => ({ media: c.media }),
});
