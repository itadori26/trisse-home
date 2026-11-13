import type { en } from './en';

/** German UI texts added to the bookmarks feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'bookmarks.sort': 'Sortierung',
  'bookmarks.sort_manual': 'Standard',
  'bookmarks.sort_az': 'A–Z',
  'bookmarks.sort_newest': 'Neueste zuerst',
};
