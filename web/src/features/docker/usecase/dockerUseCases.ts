import type { ContainerList, DockerActions } from '../domain/container';
import type { ImageUpdate, RegistryInput, UpdateSettings } from '../domain/imageUpdate';

/** Output port: docker backend. */
export interface DockerGateway {
  containers(): Promise<ContainerList>;
  actions(): Promise<DockerActions>;
  restart(name: string): Promise<void>;
  logs(name: string): Promise<string[]>;
  /** null while the update check is disabled. */
  updates(): Promise<ImageUpdate[] | null>;
  updateSettings(): Promise<UpdateSettings>;
  saveUpdateSettings(
    enabled: boolean,
    semverTags: boolean,
    registries: RegistryInput[],
  ): Promise<UpdateSettings>;
}

/** Application logic of the docker feature. */
export class DockerUseCases {
  private actionsPromise: Promise<DockerActions> | null = null;

  constructor(private readonly gateway: DockerGateway) {}

  containers(): Promise<ContainerList> {
    return this.gateway.containers();
  }

  /** Server configuration, loaded once per page view. */
  actions(): Promise<DockerActions> {
    this.actionsPromise ??= this.gateway.actions().catch((err) => {
      this.actionsPromise = null;
      throw err;
    });
    return this.actionsPromise;
  }

  restart(name: string): Promise<void> {
    return this.gateway.restart(name);
  }

  logs(name: string): Promise<string[]> {
    return this.gateway.logs(name);
  }

  /** Image update hints (#107); null while disabled. */
  updates(): Promise<ImageUpdate[] | null> {
    return this.gateway.updates();
  }

  updateSettings(): Promise<UpdateSettings> {
    return this.gateway.updateSettings();
  }

  saveUpdateSettings(
    enabled: boolean,
    semverTags: boolean,
    registries: RegistryInput[],
  ): Promise<UpdateSettings> {
    const cleaned = registries
      .map((r) => ({ host: r.host.trim().toLowerCase(), username: r.username.trim(), password: r.password }))
      .filter((r) => r.host !== '' || r.username !== '' || r.password !== '');
    return this.gateway.saveUpdateSettings(enabled, semverTags, cleaned);
  }
}
