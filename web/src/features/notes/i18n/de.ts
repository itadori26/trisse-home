import type { en } from './en';

/** German UI texts added to the notes feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'notes.word_count_one': '{n} Wort',
  'notes.word_count_other': '{n} Wörter',
  'notes.char_count_one': '{n} Zeichen',
  'notes.char_count_other': '{n} Zeichen',
};
