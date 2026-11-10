import type { HttpClient } from '../../../../shared/api-client/http';
import type { ListDTO } from '../../../../shared/api-client/generated/docker';
import type { ContainerList } from '../../domain/container';
import type { DockerGateway } from '../../usecase/dockerUseCases';

/** DockerGateway backed by the REST API. */
export class DockerApi implements DockerGateway {
  constructor(private readonly http: HttpClient) {}

  async containers(): Promise<ContainerList> {
    const d = await this.http.get<ListDTO>('/api/docker/containers');
    return {
      containers: d.containers.map((c) => ({
        id: c.id,
        name: c.name,
        image: c.image,
        state: c.state,
        status: c.status,
        health: c.health,
        project: c.project,
      })),
      total: d.total,
      running: d.running,
      stopped: d.stopped,
      unhealthy: d.unhealthy,
      error: d.error,
    };
  }
}
