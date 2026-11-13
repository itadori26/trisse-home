import type { QuickCommand } from '../features/quickactions';
import type { MessageKey } from '../shared/i18n';
import type { QuickGroup, QuickItem } from './quickSearch';

/**
 * Rows of the search panel below the hero field (#222), grouped like the
 * mockup: QUICK ADD, the result kinds, ACTIONS, WEB. Pure functions, so the
 * order and the keyboard rules are unit-tested; the panel only renders them.
 */
export type Row =
  /** Add something from a capture prefix; shortcut is shown on the row. */
  | { kind: 'capture'; id: string; command: QuickCommand; shortcut: 'Enter' | 'Ctrl Enter' }
  | { kind: 'result'; id: string; item: QuickItem }
  /** Restart a container or wake a device ("restart nextcloud"). */
  | { kind: 'action'; id: string; command: QuickCommand }
  | { kind: 'web'; id: string };

export interface RowGroup {
  id: string;
  label: MessageKey;
  rows: Row[];
}

export interface RowInput {
  /** What the text means as a command, e.g. "t Milch" → to-do. */
  command: QuickCommand | null;
  /** The same text added as the other kind (to-do ↔ note), for Ctrl+Enter. */
  alternative: QuickCommand | null;
  /** Classic results (bookmarks, services, notes, to-dos). */
  groups: QuickGroup[];
  /** Whether a web search row is offered (search module enabled, text typed). */
  web: boolean;
}

const isCapture = (c: QuickCommand) => c.kind === 'todo' || c.kind === 'note' || c.kind === 'bookmark';

/** The groups in display order; empty groups are left out. */
export function buildRows(input: RowInput): RowGroup[] {
  const out: RowGroup[] = [];
  const { command, alternative } = input;
  if (command && isCapture(command)) {
    const rows: Row[] = [{ kind: 'capture', id: 'capture', command, shortcut: 'Enter' }];
    if (alternative && isCapture(alternative))
      rows.push({ kind: 'capture', id: 'capture-alt', command: alternative, shortcut: 'Ctrl Enter' });
    out.push({ id: 'quick', label: 'app.group.quick_add', rows });
  }
  // A command replaces the classic results, as in the palette before (#67).
  if (!command) {
    for (const g of input.groups) {
      out.push({
        id: g.kind,
        label: g.label,
        rows: g.items.map((item) => ({ kind: 'result', id: `${item.kind}:${item.id}`, item })),
      });
    }
  }
  if (command && !isCapture(command)) {
    out.push({
      id: 'actions',
      label: 'app.group.actions',
      rows: [{ kind: 'action', id: 'action', command }],
    });
  }
  if (input.web) out.push({ id: 'web', label: 'app.group.web', rows: [{ kind: 'web', id: 'web' }] });
  return out.filter((g) => g.rows.length > 0);
}

/** What a key press asks for. */
export type KeyIntent = 'down' | 'up' | 'run' | 'alternative' | 'web' | 'close' | 'none';

/**
 * Keyboard rules of the panel: arrows move, Enter runs the selected row,
 * Ctrl/Cmd+Enter adds the alternative, Alt+Enter (and Shift+Enter, #221)
 * searches the web, Esc closes.
 */
export function keyIntent(
  e: Pick<KeyboardEvent, 'key' | 'ctrlKey' | 'metaKey' | 'altKey' | 'shiftKey'>,
): KeyIntent {
  switch (e.key) {
    case 'ArrowDown':
      return 'down';
    case 'ArrowUp':
      return 'up';
    case 'Escape':
      return 'close';
    case 'Enter':
      if (e.ctrlKey || e.metaKey) return 'alternative';
      if (e.altKey || e.shiftKey) return 'web';
      return 'run';
    default:
      return 'none';
  }
}

/** Moves the selection, wrapping around. */
export function move(active: number, count: number, intent: 'down' | 'up'): number {
  if (count === 0) return 0;
  return (active + (intent === 'down' ? 1 : count - 1)) % count;
}
