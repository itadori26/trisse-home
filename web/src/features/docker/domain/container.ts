/** Health check state of a container. */
export type Health = 'none' | 'healthy' | 'unhealthy' | 'starting';

/** A Docker container. */
export interface Container {
  id: string;
  name: string;
  image: string;
  state: string;
  status: string;
  health: Health;
  project: string;
}

/** The container list with counts. */
export interface ContainerList {
  containers: Container[];
  total: number;
  running: number;
  stopped: number;
  unhealthy: number;
  error: string;
}

/** Visual status of a container. */
export type Indicator = 'ok' | 'warning' | 'error' | 'off';

/** Maps state and health to a traffic light. */
export function indicator(c: Container): Indicator {
  if (c.state !== 'running') return c.state === 'restarting' || c.state === 'dead' ? 'error' : 'off';
  if (c.health === 'unhealthy') return 'error';
  if (c.health === 'starting') return 'warning';
  return 'ok';
}

/** Configuration of the docker widget. */
export interface DockerConfig {
  /** Show only containers that need attention. */
  problemsOnly: boolean;
  /** Show only containers that are not running (#119). */
  stoppedOnly: boolean;
  /** Only containers of this compose project ('' = all). */
  project: string;
}

/** Reads a stored widget configuration. */
export function parseDockerConfig(raw: Record<string, unknown>): DockerConfig {
  return {
    problemsOnly: raw['problemsOnly'] === true,
    stoppedOnly: raw['stoppedOnly'] === true,
    project: typeof raw['project'] === 'string' ? raw['project'] : '',
  };
}

/** A container counts as stopped unless it is running, like the counters of the server. */
export function isStopped(c: Pick<Container, 'state'>): boolean {
  return c.state !== 'running';
}

/**
 * Applies the widget configuration to a list. The options combine: with both
 * "needs attention" and "stopped" a container must satisfy both.
 */
export function filterContainers(list: Container[], cfg: DockerConfig): Container[] {
  return list.filter(
    (c) =>
      (!cfg.project || c.project === cfg.project) &&
      (!cfg.problemsOnly || indicator(c) !== 'ok') &&
      (!cfg.stoppedOnly || isStopped(c)),
  );
}

/** Which containers offer restart and logs (opt-in on the server). */
export interface DockerActions {
  enabled: boolean;
  containers: string[];
}

/** Whether actions are available for the container. */
export function canAct(actions: DockerActions | null, name: string): boolean {
  return !!actions?.enabled && actions.containers.includes(name);
}
