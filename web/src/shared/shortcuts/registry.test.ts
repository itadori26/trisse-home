import { describe, expect, it } from 'vitest';
import { comboOf, isTyping, ShortcutConflictError, ShortcutRegistry } from './registry';

const key = (
  k: string,
  mods: Partial<{ ctrl: boolean; meta: boolean; alt: boolean; shift: boolean }> = {},
) => ({
  key: k,
  ctrlKey: !!mods.ctrl,
  metaKey: !!mods.meta,
  altKey: !!mods.alt,
  shiftKey: !!mods.shift,
});

const el = (tagName: string, extra: Record<string, unknown> = {}) =>
  ({ tagName, isContentEditable: false, ...extra }) as unknown as EventTarget;

describe('shortcut registry (#95)', () => {
  it('normalises combos', () => {
    expect(comboOf(key('k', { ctrl: true }))).toBe('Mod+K');
    expect(comboOf(key('K', { meta: true }))).toBe('Mod+K');
    expect(comboOf(key('Z', { ctrl: true, shift: true }))).toBe('Mod+Shift+Z');
    expect(comboOf(key('?', { shift: true }))).toBe('?');
    expect(comboOf(key('e', { alt: true }))).toBeNull();
    expect(comboOf(key('ArrowDown'))).toBeNull();
  });

  it('detects typing', () => {
    expect(isTyping(el('INPUT', { type: 'text' }))).toBe(true);
    expect(isTyping(el('INPUT', { type: 'checkbox' }))).toBe(false);
    expect(isTyping(el('TEXTAREA'))).toBe(true);
    expect(isTyping(el('DIV', { isContentEditable: true }))).toBe(true);
    expect(isTyping(el('BUTTON'))).toBe(false);
    expect(isTyping(null)).toBe(false);
  });

  it('runs shortcuts, respects typing and open dialogs', () => {
    const r = new ShortcutRegistry();
    const ran: string[] = [];
    r.register({ id: 'edit', combos: ['e'], display: ['E'], label: 'x', run: () => ran.push('edit') });
    r.register({
      id: 'search',
      combos: ['Mod+K'],
      display: ['Ctrl+K'],
      label: 'x',
      typingCombos: ['Mod+K'],
      run: () => ran.push('search'),
    });
    r.register({
      id: 'tabs',
      combos: ['1', '2'],
      display: ['1–9'],
      label: 'x',
      run: (c) => ran.push(`tab ${c}`),
    });

    expect(r.handle(key('e'), el('BODY'), false)).toBe(true);
    expect(r.handle(key('e'), el('INPUT', { type: 'text' }), false)).toBe(false);
    expect(r.handle(key('k', { ctrl: true }), el('INPUT', { type: 'text' }), false)).toBe(true);
    r.register({ id: 'undo', combos: ['/', 'Mod+Z'], display: ['/'], label: 'x', run: (c) => ran.push(c) });
    expect(r.handle(key('/'), el('INPUT', { type: 'text' }), false)).toBe(false);
    // Ctrl+Z inside a field stays the field's own undo.
    expect(r.handle(key('z', { ctrl: true }), el('INPUT', { type: 'text' }), false)).toBe(false);
    expect(r.handle(key('z', { ctrl: true }), el('BODY'), false)).toBe(true);
    expect(r.handle(key('2'), el('BODY'), false)).toBe(true);
    expect(r.handle(key('e'), el('BODY'), true)).toBe(false);
    expect(r.handle(key('x'), el('BODY'), false)).toBe(false);
    expect(ran).toEqual(['edit', 'search', 'Mod+Z', 'tab 2']);
  });

  it('rejects conflicts and can unregister', () => {
    const r = new ShortcutRegistry();
    const off = r.register({ id: 'a', combos: ['e'], display: [], label: '', run: () => {} });
    expect(() => r.register({ id: 'b', combos: ['e'], display: [], label: '', run: () => {} })).toThrow(
      ShortcutConflictError,
    );
    off();
    expect(() => r.register({ id: 'b', combos: ['e'], display: [], label: '', run: () => {} })).not.toThrow();
    expect(r.all().map((s) => s.id)).toEqual(['b']);
  });
});
