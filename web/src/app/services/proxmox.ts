import { ProxmoxApi } from '../../features/proxmox/adapter/api/proxmoxApi';
import { ProxmoxUseCases } from '../../features/proxmox/usecase/proxmoxUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    proxmox: ProxmoxUseCases;
  }
}

// Direct imports keep the settings UI out of the initial bundle (#230).
export const service = defineService('proxmox', (http) => new ProxmoxUseCases(new ProxmoxApi(http)));
