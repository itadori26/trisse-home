import { describe, expect, it } from 'vitest';
import { EditHistory, type EditCommand } from './history';

const widget = (id: string) => ({
  id,
  tabId: 't',
  kind: 'clock',
  config: { tz: 'UTC' },
  x: 0,
  y: 0,
  w: 2,
  h: 2,
});
const layout = (n: number): EditCommand => ({
  type: 'layout',
  tabId: 't',
  before: [{ id: 'a', x: 0, y: 0, w: 1, h: 1 }],
  after: [{ id: 'a', x: n, y: 0, w: 1, h: 1 }],
});

describe('edit history (#93)', () => {
  it('undoes and redoes in order', () => {
    const h = new EditHistory();
    expect(h.canUndo).toBe(false);
    h.record(layout(1));
    h.record(layout(2));
    expect(h.takeUndo()).toEqual(layout(2));
    expect(h.canRedo).toBe(true);
    expect(h.takeRedo()).toEqual(layout(2));
    expect(h.takeUndo()).toEqual(layout(2));
    expect(h.takeUndo()).toEqual(layout(1));
    expect(h.takeUndo()).toBeNull();
  });

  it('drops the redo stack on a new operation and keeps at most the limit', () => {
    const h = new EditHistory(3);
    for (let i = 0; i < 5; i++) h.record(layout(i));
    h.takeUndo();
    h.record(layout(9));
    expect(h.canRedo).toBe(false);
    const seen: EditCommand[] = [];
    let c;
    while ((c = h.takeUndo())) seen.push(c);
    expect(seen).toEqual([layout(9), layout(3), layout(2)]);
  });

  it('follows re-created widget IDs', () => {
    const h = new EditHistory();
    h.record({ type: 'delete', widget: widget('w1') });
    h.remap('w1', 'w2');
    h.remap('w2', 'w3');
    expect(h.resolve('w1')).toBe('w3');
    expect(h.resolveItems([{ id: 'w1', x: 1, y: 1, w: 1, h: 1 }])[0]?.id).toBe('w3');
    expect(h.resolve('other')).toBe('other');
  });

  it('puts a command back when applying it failed', () => {
    const h = new EditHistory();
    h.record(layout(1));
    const c = h.takeUndo();
    if (!c) throw new Error('missing');
    h.restore(c, true);
    expect(h.canUndo).toBe(true);
    expect(h.canRedo).toBe(false);
  });
});
