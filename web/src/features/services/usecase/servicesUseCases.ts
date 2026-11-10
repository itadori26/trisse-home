import { type Service, type ServiceInput, validateServiceInput } from '../domain/service';

/** Output port: services backend. */
export interface ServicesGateway {
  list(): Promise<Service[]>;
  create(input: ServiceInput): Promise<Service>;
  update(id: string, input: ServiceInput): Promise<Service>;
  delete(id: string): Promise<void>;
  reorder(ids: string[]): Promise<void>;
  checkNow(): Promise<Service[]>;
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
}
