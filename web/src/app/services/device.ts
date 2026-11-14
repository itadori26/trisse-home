import { DeviceApi } from '../../features/device/adapter/api/deviceApi';
import { DeviceUseCases } from '../../features/device/usecase/deviceUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    device: DeviceUseCases;
  }
}

export const service = defineService('device', (http) => new DeviceUseCases(new DeviceApi(http)));
