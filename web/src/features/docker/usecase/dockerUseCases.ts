import type { ContainerList, DockerActions } from '../domain/container';

/** Output port: docker backend. */
export interface DockerGateway {
  containers(): Promise<ContainerList>;
  actions(): Promise<DockerActions>;
  restart(name: string): Promise<void>;
  logs(name: string): Promise<string[]>;
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
}
