import { describe, expect, it } from 'vitest';
import {
  discard,
  emptyOutbox,
  enqueue,
  isPending,
  keepMine,
  overlay,
  type PendingChange,
  settle,
} from './outbox';

let n = 0;
const change = (
  op: PendingChange['op'],
  itemId: string,
  payload: Record<string, unknown> = {},
  baseVersion = 0,
): PendingChange => ({
  key: `k${++n}`,
  entity: 'todo',
  op,
  itemId,
  payload,
  baseVersion,
  at: n,
});

describe('outbox (#80)', () => {
  it('merges changes of the same item', () => {
    let s = enqueue(emptyOutbox(), change('create', 'a', { title: 'Brot' }));
    s = enqueue(s, change('update', 'a', { done: true }));
    expect(s.queue).toHaveLength(1);
    expect(s.queue[0]).toMatchObject({ op: 'create', payload: { title: 'Brot', done: true } });
    // create + delete: the server never sees it
    expect(enqueue(s, change('delete', 'a')).queue).toHaveLength(0);

    let u = enqueue(emptyOutbox(), change('update', 'b', { title: 'x' }, 3));
    u = enqueue(u, change('update', 'b', { title: 'y' }, 4));
    expect(u.queue).toEqual([
      expect.objectContaining({ op: 'update', baseVersion: 3, payload: { title: 'y' } }),
    ]);
    u = enqueue(u, change('delete', 'b', {}, 4));
    expect(u.queue).toEqual([expect.objectContaining({ op: 'delete', baseVersion: 3 })]);
    // Other items keep their order.
    u = enqueue(u, change('update', 'c', { title: 'z' }, 1));
    expect(u.queue.map((c) => c.itemId)).toEqual(['b', 'c']);
  });

  it('settles replay results in order', () => {
    let s = enqueue(
      enqueue(emptyOutbox(), change('update', 'a', { title: 'mine' }, 1)),
      change('update', 'b', {}, 1),
    );
    expect(settle(s, { kind: 'offline' })).toBe(s);
    s = settle(s, { kind: 'changed', serverVersion: 3 });
    expect(s.queue.map((c) => c.itemId)).toEqual(['b']);
    expect(s.conflicts).toEqual([expect.objectContaining({ kind: 'changed', serverVersion: 3 })]);
    s = settle(s, { kind: 'ok' });
    expect(s.queue).toHaveLength(0);
    const key = s.conflicts[0]?.change.key ?? '';
    // Keep mine: sent again on top of the server's version.
    const mine = keepMine(s, key);
    expect(mine.conflicts).toHaveLength(0);
    expect(mine.queue).toEqual([
      expect.objectContaining({ itemId: 'a', baseVersion: 3, payload: { title: 'mine' } }),
    ]);
    // Keep server: dropped.
    expect(discard(s, key)).toEqual({ queue: [], conflicts: [] });
    expect(settle(enqueue(emptyOutbox(), change('delete', 'x', {}, 1)), { kind: 'gone' }).queue).toHaveLength(
      0,
    );
    const rejected = settle(enqueue(emptyOutbox(), change('create', 'y', { title: '' })), {
      kind: 'rejected',
      message: 'empty',
    });
    expect(rejected.conflicts[0]).toMatchObject({ kind: 'rejected', message: 'empty' });
    expect(keepMine(rejected, rejected.conflicts[0]?.change.key ?? '')).toBe(rejected);
  });

  it('overlays waiting changes on a cached list', () => {
    const items = [
      { id: 'a', title: 'A', done: false },
      { id: 'b', title: 'B', done: false },
    ];
    let s = enqueue(emptyOutbox(), change('update', 'a', { done: true }, 1));
    s = enqueue(s, change('delete', 'b', {}, 1));
    s = enqueue(s, change('create', 'c', { title: 'C' }));
    const out = overlay(items, s, 'todo', (c) => ({
      id: c.itemId,
      title: String(c.payload['title']),
      done: false,
    }));
    expect(out).toEqual([
      { id: 'a', title: 'A', done: true },
      { id: 'c', title: 'C', done: false },
    ]);
    expect(overlay(items, s, 'note', () => ({ id: '', title: '', done: false }))).toEqual(items);
    expect(isPending(s, 'todo', 'a')).toBe(true);
    expect(isPending(s, 'todo', 'x')).toBe(false);
  });
});
