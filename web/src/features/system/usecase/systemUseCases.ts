import type { SystemSnapshot } from '../domain/system';

/** Output port: system metrics backend. */
export interface SystemGateway {
  snapshot(): Promise<SystemSnapshot>;
}

/** Application logic of the system feature. */
export class SystemUseCases {
  constructor(private readonly gateway: SystemGateway) {}

  snapshot(): Promise<SystemSnapshot> {
    return this.gateway.snapshot();
  }
}
