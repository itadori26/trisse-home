import type { ShortcutRegistry } from './registry';

export { comboOf, isTyping, ShortcutConflictError, ShortcutRegistry, type Shortcut } from './registry';

/**
 * The single global listener (adapter): forwards key presses to the
 * registry. Returns a function that detaches it.
 */
export function attachShortcuts(registry: ShortcutRegistry, target: Window = window): () => void {
  const listener = (e: KeyboardEvent) => {
    if (e.defaultPrevented) return;
    const dialogOpen = !!document.querySelector('dialog[open]');
    if (registry.handle(e, e.target, dialogOpen)) e.preventDefault();
  };
  target.addEventListener('keydown', listener);
  return () => target.removeEventListener('keydown', listener);
}
