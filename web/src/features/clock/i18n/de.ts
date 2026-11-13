import type { en } from './en';

/** German UI texts added to the clock feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'clock.show_week': 'Kalenderwoche anzeigen',
  'clock.week': 'KW {n}',
};
