import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'releases',
  label: 'settings.section.releases',
  module: 'releases',
  load: () => import('../../features/releases').then((m) => m.ReleasesSettings),
  props: (c) => ({ releases: c.releases }),
});
