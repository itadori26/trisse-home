import type { DeviceState } from '../domain/device';

/** Output port: device backend. */
export interface DeviceGateway {
  state(): Promise<DeviceState>;
  changeMode(mode: 'headless' | 'kiosk'): Promise<DeviceState>;
}

/** Display mode of the appliance (#140). */
export class DeviceUseCases {
  constructor(private readonly gateway: DeviceGateway) {}

  state(): Promise<DeviceState> {
    return this.gateway.state();
  }
  changeMode(mode: 'headless' | 'kiosk'): Promise<DeviceState> {
    return this.gateway.changeMode(mode);
  }
}
