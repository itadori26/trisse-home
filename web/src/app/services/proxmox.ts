import { ProxmoxApi, ProxmoxUseCases } from '../../features/proxmox';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    proxmox: ProxmoxUseCases;
  }
}

export const service = defineService('proxmox', (http) => new ProxmoxUseCases(new ProxmoxApi(http)));
