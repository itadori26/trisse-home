import type { Incident, ServiceUptime } from '../domain/uptime';
import { type Discovery, type DiscoverySettings, validInterval } from '../domain/discovery';
import {
  type LatencyPoint,
  type Service,
  type ServiceInput,
  validateServiceInput,
  visibleServices,
} from '../domain/service';
import { t } from '../../../shared/i18n';

/** Output port: services backend. */
export interface ServicesGateway {
  list(): Promise<Service[]>;
  create(input: ServiceInput): Promise<Service>;
  update(id: string, input: ServiceInput): Promise<Service>;
  delete(id: string): Promise<void>;
  reorder(ids: string[]): Promise<void>;
  checkNow(): Promise<Service[]>;
  history(hours: number): Promise<Record<string, LatencyPoint[]>>;
  uptime(): Promise<ServiceUptime[]>;
  incidents(serviceId: string, limit: number): Promise<Incident[]>;
  wake(id: string): Promise<void>;
  discovery(): Promise<Discovery>;
  saveDiscovery(settings: DiscoverySettings): Promise<Discovery>;
  syncDiscovery(): Promise<Discovery>;
}

/** Error for input rejected before it reaches the backend. */
export class ServiceValidationError extends Error {
  override name = 'ServiceValidationError';
}

/** Application logic of the services feature. */
export class ServicesUseCases {
  constructor(private readonly gateway: ServicesGateway) {}

  /** Tiles shown on the dashboard (without those of removed containers). */
  async list(): Promise<Service[]> {
    return visibleServices(await this.gateway.list());
  }

  /** Every tile, including hidden ones, for the settings. */
  listAll(): Promise<Service[]> {
    return this.gateway.list();
  }

  /** Creates or updates a service after validation. */
  save(input: ServiceInput, id?: string): Promise<Service> {
    const problem = validateServiceInput(input);
    if (problem) return Promise.reject(new ServiceValidationError(problem));
    return id ? this.gateway.update(id, input) : this.gateway.create(input);
  }

  delete(id: string): Promise<void> {
    return this.gateway.delete(id);
  }

  /** Moves a service by `delta` positions and persists the order. */
  async move(services: readonly Service[], id: string, delta: number): Promise<Service[]> {
    const from = services.findIndex((s) => s.id === id);
    const to = from + delta;
    if (from < 0 || to < 0 || to >= services.length) return [...services];
    const out = [...services];
    const [moved] = out.splice(from, 1);
    if (moved) out.splice(to, 0, moved);
    await this.gateway.reorder(out.map((s) => s.id));
    return out;
  }

  async checkNow(): Promise<Service[]> {
    return visibleServices(await this.gateway.checkNow());
  }

  /** Response time history of the last 24 hours per service ID. */
  history(): Promise<Record<string, LatencyPoint[]>> {
    return this.gateway.history(24);
  }

  /** Sends a Wake-on-LAN packet to the service's MAC address. */
  wake(id: string): Promise<void> {
    return this.gateway.wake(id);
  }

  /** Uptime of every service (24 h, 7, 30, 90 days and daily bars). */
  uptime(): Promise<ServiceUptime[]> {
    return this.gateway.uptime();
  }

  /** Recent outages of one service, or of all when serviceId is empty. */
  incidents(serviceId = '', limit = 20): Promise<Incident[]> {
    return this.gateway.incidents(serviceId, limit);
  }

  /** State of Docker autodiscovery (#78). */
  discovery(): Promise<Discovery> {
    return this.gateway.discovery();
  }

  /** Saves the discovery settings after validation. */
  saveDiscovery(settings: DiscoverySettings): Promise<Discovery> {
    if (!validInterval(settings.intervalMinutes))
      return Promise.reject(new ServiceValidationError(t('services.discovery.invalid_interval')));
    return this.gateway.saveDiscovery(settings);
  }

  /** Reads the container labels now and updates the tiles. */
  syncDiscovery(): Promise<Discovery> {
    return this.gateway.syncDiscovery();
  }
}
