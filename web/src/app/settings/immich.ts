import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'immich',
  label: 'settings.section.immich',
  module: 'immich',
  load: () => import('../../features/immich').then((m) => m.ImmichSettings),
  props: (c) => ({ immich: c.immich }),
});
