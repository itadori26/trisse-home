import { describe, expect, it } from 'vitest';
import { type Container, filterContainers, indicator, canAct } from './container';

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
    expect(filterContainers(list, { problemsOnly: true, project: 'x' }).map((x) => x.name)).toEqual(['b']);
    expect(filterContainers(list, { problemsOnly: false, project: '' })).toHaveLength(3);
  });
  it('offers actions only for allowed containers', () => {
    expect(canAct({ enabled: true, containers: ['homarr'] }, 'homarr')).toBe(true);
    expect(canAct({ enabled: true, containers: ['homarr'] }, 'redis')).toBe(false);
    expect(canAct({ enabled: false, containers: ['homarr'] }, 'homarr')).toBe(false);
    expect(canAct(null, 'homarr')).toBe(false);
  });
});
