import type { en } from './en';

/** German UI texts added to the docker feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'docker.only_stopped': 'Nur gestoppte Container anzeigen',
};
