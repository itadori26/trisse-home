import type { Disk } from '../domain/disk';

/** Output port: disk health backend. */
export interface DiskHealthGateway {
  disks(): Promise<Disk[]>;
}

/** Application logic of the disk health feature. */
export class DiskHealthUseCases {
  constructor(private readonly gateway: DiskHealthGateway) {}

  disks(): Promise<Disk[]> {
    return this.gateway.disks();
  }
}
