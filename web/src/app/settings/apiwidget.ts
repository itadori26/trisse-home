import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'apiwidget',
  label: 'settings.section.apiwidget',
  module: 'apiwidget',
  load: () => import('../../features/apiwidget').then((m) => m.ApiWidgetSettings),
  props: (c) => ({ apiwidget: c.apiwidget }),
});
