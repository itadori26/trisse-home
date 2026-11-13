import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'proxmox',
  label: 'settings.section.proxmox',
  module: 'proxmox',
  load: () => import('../../features/proxmox').then((m) => m.ProxmoxSettings),
  props: (c) => ({ proxmox: c.proxmox }),
});
