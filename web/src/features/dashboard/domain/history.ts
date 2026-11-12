import type { Widget, WidgetConfig } from './dashboard';
import type { LayoutItem } from './layout';

/** Maximum number of undoable steps (#93). */
export const HISTORY_LIMIT = 50;

/** An undoable edit-mode operation, stored as data. */
export type EditCommand =
  | { type: 'layout'; tabId: string; before: LayoutItem[]; after: LayoutItem[] }
  /** Added or duplicated widget: undo deletes it, redo re-creates it. */
  | { type: 'add'; widget: Widget }
  /** Deleted widget with its configuration: undo re-creates it. */
  | { type: 'delete'; widget: Widget }
  | { type: 'configure'; widgetId: string; before: WidgetConfig; after: WidgetConfig };

/**
 * Undo/redo stack (command pattern). Re-creating a deleted widget gives it a
 * new ID, so the history keeps a map from old to new IDs and resolves the
 * IDs of every command through it.
 */
export class EditHistory {
  private undoStack: EditCommand[] = [];
  private redoStack: EditCommand[] = [];
  private readonly ids = new Map<string, string>();

  constructor(private readonly limit = HISTORY_LIMIT) {}

  /** Records a new operation; it clears the redo stack. */
  record(c: EditCommand): void {
    this.undoStack.push(c);
    if (this.undoStack.length > this.limit) this.undoStack.shift();
    this.redoStack = [];
  }

  get canUndo(): boolean {
    return this.undoStack.length > 0;
  }

  get canRedo(): boolean {
    return this.redoStack.length > 0;
  }

  /** Takes the command to revert, or null. */
  takeUndo(): EditCommand | null {
    const c = this.undoStack.pop() ?? null;
    if (c) this.redoStack.push(c);
    return c;
  }

  /** Takes the command to apply again, or null. */
  takeRedo(): EditCommand | null {
    const c = this.redoStack.pop() ?? null;
    if (c) this.undoStack.push(c);
    return c;
  }

  /** Puts a command back after a failed undo/redo. */
  restore(c: EditCommand, wasUndo: boolean): void {
    if (wasUndo) {
      this.redoStack.pop();
      this.undoStack.push(c);
    } else {
      this.undoStack.pop();
      this.redoStack.push(c);
    }
  }

  /** Notes that a widget was re-created under a new ID. */
  remap(oldId: string, newId: string): void {
    this.ids.set(oldId, newId);
  }

  /** The current ID of a widget, following re-creations. */
  resolve(id: string): string {
    let cur = id;
    for (let i = 0; i < 100 && this.ids.has(cur); i++) cur = this.ids.get(cur) ?? cur;
    return cur;
  }

  /** Layout items with resolved IDs. */
  resolveItems(items: readonly LayoutItem[]): LayoutItem[] {
    return items.map((i) => ({ ...i, id: this.resolve(i.id) }));
  }

  /** Clears everything (e.g. when leaving edit mode is not required; kept per session). */
  clear(): void {
    this.undoStack = [];
    this.redoStack = [];
    this.ids.clear();
  }
}
