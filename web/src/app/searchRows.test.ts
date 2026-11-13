import { describe, expect, it } from 'vitest';
import type { QuickCommand } from '../features/quickactions';
import type { QuickGroup } from './quickSearch';
import { buildRows, keyIntent, move } from './searchRows';

const todo: QuickCommand = { kind: 'todo', title: 'buy', due: '' };
const note: QuickCommand = { kind: 'note', title: 'buy' };
const restart: QuickCommand = { kind: 'restart', target: 'nextcloud' };
const bookmarks: QuickGroup = {
  kind: 'bookmark',
  label: 'app.group.bookmarks',
  items: [{ kind: 'bookmark', id: 'b1', title: 'Buy Me a Coffee', detail: 'Projects', url: 'https://x' }],
};
const ids = (g: ReturnType<typeof buildRows>) => g.map((x) => `${x.id}:${x.rows.map((r) => r.id).join(',')}`);

describe('buildRows (#222)', () => {
  it('orders a capture like the mockup: quick add with the alternative, then web', () => {
    const g = buildRows({ command: todo, alternative: note, groups: [bookmarks], web: true });
    expect(ids(g)).toEqual(['quick:capture,capture-alt', 'web:web']);
    const [first, alt] = g[0]?.rows ?? [];
    expect(first).toMatchObject({ command: todo, shortcut: 'Enter' });
    expect(alt).toMatchObject({ command: note, shortcut: 'Ctrl Enter' });
  });
  it('shows the results and the web row for a plain search', () => {
    const g = buildRows({ command: null, alternative: null, groups: [bookmarks], web: true });
    expect(ids(g)).toEqual(['bookmark:bookmark:b1', 'web:web']);
  });
  it('puts restart and wake under actions', () => {
    const g = buildRows({ command: restart, alternative: null, groups: [bookmarks], web: true });
    expect(ids(g)).toEqual(['actions:action', 'web:web']);
  });
  it('leaves out the web row without the search module and empty groups', () => {
    expect(
      buildRows({ command: null, alternative: null, groups: [{ ...bookmarks, items: [] }], web: false }),
    ).toEqual([]);
  });
});

describe('keyIntent (#222)', () => {
  const k = (
    key: string,
    mods: Partial<Record<'ctrlKey' | 'metaKey' | 'altKey' | 'shiftKey', boolean>> = {},
  ) => keyIntent({ key, ctrlKey: false, metaKey: false, altKey: false, shiftKey: false, ...mods });
  it('maps the keys of the mockup', () => {
    expect(k('Enter')).toBe('run');
    expect(k('Enter', { ctrlKey: true })).toBe('alternative');
    expect(k('Enter', { metaKey: true })).toBe('alternative');
    expect(k('Enter', { altKey: true })).toBe('web');
    expect(k('Enter', { shiftKey: true })).toBe('web');
    expect(k('Escape')).toBe('close');
    expect(k('ArrowDown')).toBe('down');
    expect(k('ArrowUp')).toBe('up');
    expect(k('a')).toBe('none');
  });
  it('wraps the selection', () => {
    expect(move(0, 3, 'down')).toBe(1);
    expect(move(2, 3, 'down')).toBe(0);
    expect(move(0, 3, 'up')).toBe(2);
    expect(move(0, 0, 'down')).toBe(0);
  });
});
