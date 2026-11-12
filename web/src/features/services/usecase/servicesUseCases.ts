import type { Incident, ServiceUptime } from '../domain/uptime';
import { type LatencyPoint, type Service, type ServiceInput, validateServiceInput } from '../domain/service';

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
}

/** Error for input rejected before it reaches the backend. */
export class ServiceValidationError extends Error {
  override name = 'ServiceValidationError';
}

/** Application logic of the services feature. */
export class ServicesUseCases {
  constructor(private readonly gateway: ServicesGateway) {}

  list(): Promise<Service[]> {
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

  checkNow(): Promise<Service[]> {
    return this.gateway.checkNow();
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
}
