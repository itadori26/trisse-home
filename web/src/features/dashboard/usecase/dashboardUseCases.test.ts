import { describe, expect, it, vi } from 'vitest';
import type { Tab } from '../domain/dashboard';
import { DashboardUseCases, DashboardValidationError } from './dashboardUseCases';
import type { DashboardGateway } from './ports';

const tab = (id: string, position: number, widgets: Tab['widgets'] = []): Tab => ({
  id,
  name: id,
  position,
  background: '',
  widgets,
});

function setup() {
  const gateway = {
    load: vi.fn().mockResolvedValue([tab('b', 1), tab('a', 0)]),
    createTab: vi.fn(),
    renameTab: vi.fn(),
    deleteTab: vi.fn(),
    reorderTabs: vi.fn().mockResolvedValue(undefined),
    addWidget: vi.fn(),
    configureWidget: vi.fn(),
    saveLayout: vi.fn().mockResolvedValue(undefined),
    deleteWidget: vi.fn(),
    setBackground: vi.fn(),
    listTemplates: vi.fn(),
    saveTemplate: vi.fn(),
    deleteTemplate: vi.fn(),
    createTabFromTemplate: vi.fn(),
  } satisfies DashboardGateway;
  return { uc: new DashboardUseCases(gateway), gateway };
}

describe('DashboardUseCases', () => {
  it('sorts tabs by position', async () => {
    const { uc } = setup();
    expect((await uc.load()).map((t) => t.id)).toEqual(['a', 'b']);
  });

  it('restores a deleted widget at its old place or the first free spot (#93)', async () => {
    const { uc, gateway } = setup();
    const gone = { id: 'w2', tabId: 'a', kind: 'notes', config: { id: 'n1' }, x: 4, y: 0, w: 4, h: 3 };
    await uc.restoreWidget(tab('a', 0), gone);
    expect(gateway.addWidget).toHaveBeenLastCalledWith(
      'a',
      'notes',
      { x: 4, y: 0, w: 4, h: 3 },
      { id: 'n1' },
    );
    const blocker = { id: 'w3', tabId: 'a', kind: 'clock', config: {}, x: 4, y: 0, w: 4, h: 2 };
    await uc.restoreWidget(tab('a', 0, [blocker]), gone);
    expect(gateway.addWidget).toHaveBeenLastCalledWith(
      'a',
      'notes',
      { x: 0, y: 0, w: 4, h: 3 },
      { id: 'n1' },
    );
  });

  it('rejects invalid tab names without calling the backend', async () => {
    const { uc, gateway } = setup();
    await expect(uc.createTab('  ')).rejects.toBeInstanceOf(DashboardValidationError);
    expect(gateway.createTab).not.toHaveBeenCalled();
  });

  it('places new widgets at the first free spot', async () => {
    const { uc, gateway } = setup();
    const t = tab('a', 0, [{ id: 'w', tabId: 'a', kind: 'clock', config: {}, x: 0, y: 0, w: 4, h: 2 }]);
    await uc.addWidget(t, {
      kind: 'weather',
      title: 'Weather',
      description: '',
      defaultSize: { w: 4, h: 3 },
      category: 'information',
      keywords: [],
    });
    expect(gateway.addWidget).toHaveBeenCalledWith('a', 'weather', { x: 4, y: 0, w: 4, h: 3 }, {});
  });

  it('skips saving unchanged layouts', async () => {
    const { uc, gateway } = setup();
    const items = [{ id: 'w', x: 0, y: 0, w: 1, h: 1 }];
    await uc.saveLayout('a', items, [...items]);
    expect(gateway.saveLayout).not.toHaveBeenCalled();
    await uc.saveLayout('a', items, [{ id: 'w', x: 1, y: 0, w: 1, h: 1 }]);
    expect(gateway.saveLayout).toHaveBeenCalledOnce();
  });

  it('reorders tabs and renumbers positions', async () => {
    const { uc, gateway } = setup();
    const out = await uc.moveTab([tab('a', 0), tab('b', 1), tab('c', 2)], 2, 0);
    expect(out.map((t) => [t.id, t.position])).toEqual([
      ['c', 0],
      ['a', 1],
      ['b', 2],
    ]);
    expect(gateway.reorderTabs).toHaveBeenCalledWith(['c', 'a', 'b']);
  });

  it('duplicates a widget into the first free spot', async () => {
    const { uc, gateway } = setup();
    const w = { id: 'w', tabId: 'a', kind: 'clock', config: { tz: 'UTC' }, x: 0, y: 0, w: 4, h: 2 };
    await uc.duplicateWidget(tab('a', 0, [w]), w);
    expect(gateway.addWidget).toHaveBeenCalledWith('a', 'clock', { x: 4, y: 0, w: 4, h: 2 }, { tz: 'UTC' });
  });

  it('validates backgrounds and template names', async () => {
    const { uc, gateway } = setup();
    await expect(uc.setBackground('a', 'ftp://x')).rejects.toBeInstanceOf(DashboardValidationError);
    await uc.setBackground('a', ' https://img.example/a.jpg ');
    expect(gateway.setBackground).toHaveBeenCalledWith('a', 'https://img.example/a.jpg');
    await expect(uc.saveTemplate('a', ' ')).rejects.toBeInstanceOf(DashboardValidationError);
    await uc.createTabFromTemplate('t', '');
    expect(gateway.createTabFromTemplate).toHaveBeenCalledWith('t', '');
  });
});
