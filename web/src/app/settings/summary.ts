import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'summary',
  label: 'settings.section.summary',
  module: 'summary',
  load: () => import('../../features/summary/ui/SummarySettings.svelte').then((m) => m.default),
  props: (c) => ({ summary: c.summary }),
});
