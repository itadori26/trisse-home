/** A Markdown note. */
export interface Note {
  id: string;
  title: string;
  body: string;
  pinned: boolean;
  updatedAt: Date;
  /** Stored version (#80); 0 = only on this device so far (offline). */
  version: number;
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

/** A piece of highlighted search output; hit marks a match (#73). */
export interface Segment {
  text: string;
  hit: boolean;
}

/** A note found by the full-text search. */
export interface SearchHit {
  id: string;
  title: Segment[];
  snippet: Segment[];
}

/** A note that links to another one. */
export interface NoteRef {
  id: string;
  title: string;
}

/** An earlier version of a note. */
export interface Revision {
  id: string;
  title: string;
  createdAt: Date;
  size: number;
}

/** One line of a diff from a revision to the current note. */
export interface DiffLine {
  op: ' ' | '+' | '-';
  text: string;
}

/** A revision with its text and the diff to the current note. */
export interface RevisionDetail extends Revision {
  body: string;
  diff: DiffLine[];
}

/** Prefix of internal links to other notes, e.g. "#note:Shopping%20list". */
export const NOTE_LINK_PREFIX = '#note:';

const CODE = /(```[\s\S]*?```|~~~[\s\S]*?~~~|`[^`\n]*`)/;
const WIKI = /\[\[([^[\]|\n]+)(?:\|([^[\]\n]*))?\]\]/g;

/**
 * Turns [[Title]] and [[Title|label]] into Markdown links to the note with
 * that title; code spans and blocks stay untouched (mirrors the backend).
 */
export function linkWikiLinks(body: string): string {
  return body
    .split(CODE)
    .map((part, i) =>
      i % 2 === 1
        ? part
        : part.replace(WIKI, (_, title: string, label?: string) => {
            const name = title.trim();
            const text = (label?.trim() || name).replace(/[[\]]/g, '');
            return `[${text}](${NOTE_LINK_PREFIX}${encodeURIComponent(name)})`;
          }),
    )
    .join('');
}

/** Returns the note title of an internal link, or null for other links. */
export function noteLinkTitle(href: string): string | null {
  if (!href.startsWith(NOTE_LINK_PREFIX)) return null;
  try {
    return decodeURIComponent(href.slice(NOTE_LINK_PREFIX.length));
  } catch {
    return null;
  }
}

/** Finds the note a link points to: same title, ignoring case and spaces. */
export function findByTitle<T extends { title: string }>(notes: readonly T[], title: string): T | undefined {
  const want = title.trim().toLowerCase();
  return notes.find((n) => n.title.trim().toLowerCase() === want);
}

/** Plain text of highlighted segments. */
export const plain = (segments: readonly Segment[]): string => segments.map((s) => s.text).join('');

/** Length of a note's text (#121). */
export interface TextCount {
  words: number;
  /** Visible characters including spaces; line breaks are not counted. */
  characters: number;
}

const segmenter = typeof Intl !== 'undefined' && 'Segmenter' in Intl ? new Intl.Segmenter() : undefined;

/**
 * Counts the words and characters of a note.
 *
 * Markdown syntax does not count as words: code fence lines are skipped, a
 * link `[text](url)` or image counts only its text, and a token needs a letter
 * or digit, so `#`, `-`, `---`, `>` and `*` stand for nothing. Characters are
 * grapheme clusters, so an emoji, a flag or a letter with an accent is one.
 */
export function countText(text: string): TextCount {
  const prose = text
    .split(/\r?\n/)
    .filter((line) => !/^\s*(```|~~~)/.test(line))
    .join('\n')
    .replace(/!?\[([^\]]*)\]\([^)]*\)/g, '$1');
  const words = prose.split(/\s+/).filter((token) => /[\p{L}\p{N}]/u.test(token)).length;
  const visible = text.replace(/[\r\n]/g, '');
  const characters = segmenter ? [...segmenter.segment(visible)].length : [...visible].length;
  return { words, characters };
}
