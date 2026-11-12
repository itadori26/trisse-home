import { ApiError, type HttpClient } from '../../../../shared/api-client/http';
import type {
  CapabilitiesDTO,
  ImageUpdateDTO,
  ListDTO,
  LogsDTO,
  UpdateSettingsDTO,
} from '../../../../shared/api-client/generated/docker';
import type { UpdateSettingsInputDTO } from '../../../../shared/api-client/generated/docker-requests';
import type { ImageUpdate, RegistryInput, UpdateSettings } from '../../domain/imageUpdate';
import type { ContainerList, DockerActions } from '../../domain/container';
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

  async actions(): Promise<DockerActions> {
    const d = await this.http.get<CapabilitiesDTO>('/api/docker/actions');
    return { enabled: d.enabled, containers: d.containers };
  }

  restart(name: string): Promise<void> {
    return this.http.post(`/api/docker/containers/${encodeURIComponent(name)}/restart`);
  }

  async logs(name: string): Promise<string[]> {
    return (await this.http.get<LogsDTO>(`/api/docker/containers/${encodeURIComponent(name)}/logs`)).lines;
  }

  async updates(): Promise<ImageUpdate[] | null> {
    try {
      const list = await this.http.get<ImageUpdateDTO[]>('/api/docker/updates');
      return list.map((d) => ({
        ...d,
        containers: d.containers ?? [],
        checkedAt: d.checkedAt && !d.checkedAt.startsWith('0001') ? new Date(d.checkedAt) : null,
      }));
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) return null;
      throw err;
    }
  }

  updateSettings(): Promise<UpdateSettings> {
    return this.http.get<UpdateSettingsDTO>('/api/docker/updates/settings');
  }

  saveUpdateSettings(
    enabled: boolean,
    semverTags: boolean,
    registries: RegistryInput[],
  ): Promise<UpdateSettings> {
    const body: UpdateSettingsInputDTO = { enabled, semverTags, registries };
    return this.http.put<UpdateSettingsDTO>('/api/docker/updates/settings', body);
  }
}
