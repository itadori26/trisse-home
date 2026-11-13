import { describe, expect, it } from 'vitest';
import {
  type Container,
  canAct,
  filterContainers,
  indicator,
  isStopped,
  parseDockerConfig,
} from './container';

const c = (name: string, state: string, health: Container['health'] = 'none', project = ''): Container => ({
  id: name,
  name,
  image: 'img',
  state,
  status: '',
  health,
  project,
});

describe('docker domain', () => {
  it('maps indicators', () => {
    expect(indicator(c('a', 'running'))).toBe('ok');
    expect(indicator(c('a', 'running', 'unhealthy'))).toBe('error');
    expect(indicator(c('a', 'running', 'starting'))).toBe('warning');
    expect(indicator(c('a', 'exited'))).toBe('off');
    expect(indicator(c('a', 'restarting'))).toBe('error');
  });
  it('filters by project and problems', () => {
    const list = [
      c('a', 'running', 'none', 'x'),
      c('b', 'exited', 'none', 'x'),
      c('c', 'exited', 'none', 'y'),
    ];
    expect(
      filterContainers(list, { problemsOnly: true, stoppedOnly: false, project: 'x' }).map((x) => x.name),
    ).toEqual(['b']);
    expect(filterContainers(list, { problemsOnly: false, stoppedOnly: false, project: '' })).toHaveLength(3);
  });
  it('offers actions only for allowed containers', () => {
    expect(canAct({ enabled: true, containers: ['homarr'] }, 'homarr')).toBe(true);
    expect(canAct({ enabled: true, containers: ['homarr'] }, 'redis')).toBe(false);
    expect(canAct({ enabled: false, containers: ['homarr'] }, 'homarr')).toBe(false);
    expect(canAct(null, 'homarr')).toBe(false);
  });
});

describe('stopped only (#119)', () => {
  const list = [
    c('web', 'running'),
    c('sick', 'running', 'unhealthy'),
    c('done', 'exited'),
    c('loop', 'restarting'),
    c('made', 'created', 'none', 'x'),
  ];
  const names = (cfg: Parameters<typeof filterContainers>[1]) =>
    filterContainers(list, cfg).map((x) => x.name);
  const off = { problemsOnly: false, stoppedOnly: false, project: '' };

  it('is off by default and read from the stored config', () => {
    expect(parseDockerConfig({}).stoppedOnly).toBe(false);
    expect(parseDockerConfig({ stoppedOnly: true }).stoppedOnly).toBe(true);
    expect(parseDockerConfig({ stoppedOnly: 'yes' }).stoppedOnly).toBe(false);
  });

  it('counts everything that is not running as stopped', () => {
    expect(list.filter(isStopped).map((x) => x.name)).toEqual(['done', 'loop', 'made']);
  });

  it('shows only stopped containers when enabled', () => {
    expect(names(off)).toHaveLength(5);
    expect(names({ ...off, stoppedOnly: true })).toEqual(['done', 'loop', 'made']);
  });

  it('combines with the other options', () => {
    expect(names({ ...off, stoppedOnly: true, project: 'x' })).toEqual(['made']);
    // An unhealthy container runs, so it is not stopped even though it needs attention.
    expect(names({ ...off, stoppedOnly: true, problemsOnly: true })).toEqual(['done', 'loop', 'made']);
  });
});
