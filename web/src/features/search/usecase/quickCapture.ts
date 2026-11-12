/**
 * Output port of the search bar for quick capture (#105): "t Milch kaufen"
 * creates a to-do instead of searching. Implemented in the app layer, so
 * the search feature does not know to-dos, notes or bookmarks.
 */
export interface QuickCapture {
  /** What Enter will do, or null when the input is a normal search. */
  preview(input: string): string | null;
  /** Creates the item; null when the input is a normal search. */
  run(input: string): Promise<{ message: string; undo?: () => Promise<void>; openUrl?: string } | null>;
}
