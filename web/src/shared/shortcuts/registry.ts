/**
 * Keyboard shortcut registry (#95): one place that knows every global
 * shortcut, used by the single global listener and by the "?" overview.
 * Only plain keys (no Alt/Ctrl/Meta) and Ctrl/Cmd+K are used, so browser
 * and screen reader shortcuts keep working.
 */

/** The parts of a keyboard event the registry looks at. */
export type KeyInput = Pick<KeyboardEvent, 'key' | 'ctrlKey' | 'metaKey' | 'altKey' | 'shiftKey'>;

/** A registered shortcut. */
export interface Shortcut {
  /** Unique command id, e.g. "search.open". */
  id: string;
  /** Normalised key combinations, e.g. ["Mod+K", "/"] or ["1", …, "9"]. */
  combos: string[];
  /** What the overview shows for the keys, e.g. "1–9". */
  display: string[];
  /** Translation key of the description. */
  label: string;
  /** Also active while the user types in a field (only for Mod+ combos). */
  whileTyping?: boolean;
  run: (combo: string) => void;
}

/**
 * Normalises an event to a combo string: "Mod+K" for Ctrl or Cmd plus a
 * letter, the plain key otherwise (Shift is ignored for printable keys, so
 * "?" works on every layout); null for Alt combinations and other modifiers.
 */
export function comboOf(e: KeyInput): string | null {
  if (e.altKey) return null;
  if (e.ctrlKey || e.metaKey) return e.key.length === 1 ? `Mod+${e.key.toUpperCase()}` : null;
  return e.key.length === 1 ? e.key : null;
}

/** Whether the event comes from a place where the user types. */
export function isTyping(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null;
  if (!el || typeof el.tagName !== 'string') return false;
  if (el.isContentEditable) return true;
  if (el.tagName === 'TEXTAREA' || el.tagName === 'SELECT') return true;
  if (el.tagName !== 'INPUT') return false;
  const type = (el as HTMLInputElement).type;
  return !['checkbox', 'radio', 'button', 'submit', 'reset', 'range', 'color', 'file'].includes(type);
}

/** Raised when two shortcuts claim the same combo. */
export class ShortcutConflictError extends Error {}

/** The registry. */
export class ShortcutRegistry {
  private readonly byCombo = new Map<string, Shortcut>();
  private readonly list: Shortcut[] = [];

  /** Registers a shortcut; returns a function that removes it. */
  register(s: Shortcut): () => void {
    for (const c of s.combos) {
      const other = this.byCombo.get(c);
      if (other) throw new ShortcutConflictError(`${c} is used by ${other.id} and ${s.id}`);
    }
    for (const c of s.combos) this.byCombo.set(c, s);
    this.list.push(s);
    return () => {
      for (const c of s.combos) if (this.byCombo.get(c) === s) this.byCombo.delete(c);
      const i = this.list.indexOf(s);
      if (i >= 0) this.list.splice(i, 1);
    };
  }

  /** All shortcuts in registration order (for the overview). */
  all(): readonly Shortcut[] {
    return this.list;
  }

  /**
   * Runs the matching shortcut and reports whether one ran. Plain keys are
   * ignored while typing; nothing runs while a modal dialog is open.
   */
  handle(e: KeyInput, target: EventTarget | null, dialogOpen: boolean): boolean {
    const combo = comboOf(e);
    if (!combo || dialogOpen) return false;
    const s = this.byCombo.get(combo);
    if (!s) return false;
    // Mod+ combos never type text, so they also work inside fields.
    if (isTyping(target) && !s.whileTyping && !combo.startsWith('Mod+')) return false;
    s.run(combo);
    return true;
  }
}
