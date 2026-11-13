import { type ArrInstance, type ArrInstanceInput, type ArrOverview, validateInstance } from '../domain/arr';

/** Output port: *arr backend. */
export interface ArrGateway {
  instances(): Promise<ArrInstance[]>;
  add(input: ArrInstanceInput): Promise<ArrInstance>;
  update(id: string, input: ArrInstanceInput): Promise<ArrInstance>;
  delete(id: string): Promise<void>;
  overview(): Promise<ArrOverview>;
}

/** Application logic of the Sonarr/Radarr/Prowlarr feature (#86). */
export class ArrUseCases {
  constructor(private readonly gateway: ArrGateway) {}

  instances(): Promise<ArrInstance[]> {
    return this.gateway.instances();
  }

  /** Adds or updates an instance after validation; an empty key keeps the stored one. */
  save(input: ArrInstanceInput, id?: string): Promise<ArrInstance> {
    const problem = validateInstance(input, !id);
    if (problem) return Promise.reject(new Error(problem));
    const clean = { ...input, name: input.name.trim(), url: input.url.trim(), apiKey: input.apiKey.trim() };
    return id ? this.gateway.update(id, clean) : this.gateway.add(clean);
  }

  delete(id: string): Promise<void> {
    return this.gateway.delete(id);
  }

  overview(): Promise<ArrOverview> {
    return this.gateway.overview();
  }
}
