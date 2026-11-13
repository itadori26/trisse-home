import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The proxmox widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'proxmox',
      category: 'server',
      keywords: ['pve', 'vm', 'lxc', 'virtualisation', 'hypervisor'],
      get title() {
        return t('widget.proxmox.title');
      },
      get description() {
        return t('widget.proxmox.description');
      },
      defaultSize: { w: 4, h: 6 },
    },
    icon: 'server',
    view: lazy(
      () => import('../../features/proxmox/ui/ProxmoxWidget.svelte'),
      (config, c) => ({ config, proxmox: c.proxmox }),
    ),
    settings: lazy(
      () => import('../../features/proxmox/ui/ProxmoxWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
