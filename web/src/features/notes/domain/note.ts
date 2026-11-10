/** A Markdown note. */
export interface Note {
  id: string;
  title: string;
  body: string;
  pinned: boolean;
  updatedAt: Date;
}

/** Editable fields of a note. */
export interface NoteInput {
  title: string;
  body: string;
  pinned: boolean;
}

/** Which notes the widget shows. */
export type NotesScope = 'pinned' | 'all';

/** Configuration of the notes widget. */
export interface NotesConfig {
  scope: NotesScope;
}

/** Reads a stored widget configuration. */
export function parseNotesConfig(raw: Record<string, unknown>): NotesConfig {
  return { scope: raw['scope'] === 'all' ? 'all' : 'pinned' };
}

/** A short plain-text preview of the Markdown body. */
export function excerpt(body: string, max = 140): string {
  const text = body
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/[#>*_`~[\]()!-]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
  return text.length > max ? `${text.slice(0, max - 1)}…` : text;
}
