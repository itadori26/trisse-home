import type { ContainerList } from '../domain/container';

/** Output port: docker backend. */
export interface DockerGateway {
  containers(): Promise<ContainerList>;
}

/** Application logic of the docker feature. */
export class DockerUseCases {
  constructor(private readonly gateway: DockerGateway) {}

  containers(): Promise<ContainerList> {
    return this.gateway.containers();
  }
}
