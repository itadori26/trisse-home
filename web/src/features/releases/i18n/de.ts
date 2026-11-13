import type { en } from './en';

/** German UI texts of the releases feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.releases.title': 'Neuerscheinungen',
  'widget.releases.description': 'Neue Alben, Singles und EPs der Künstler, denen du folgst (MusicBrainz).',
  'settings.section.releases': 'Neuerscheinungen',
  'releases.none':
    'Keine Neuerscheinungen in den letzten {weeks} Wochen. Folge Künstlern unter Einstellungen → Neuerscheinungen.',
  'releases.weeks': 'Wochen',
  'releases.hint':
    'Folge Künstlern, um ihre neuen Alben, Singles und EPs zu sehen. Neuerscheinungen werden alle sechs Stunden geprüft.',
  'releases.search': 'Künstler suchen',
  'releases.follow': 'Folgen',
  'releases.followed': 'Gefolgte Künstler ({count})',
  'releases.unfollow': '{name} entfolgen',
  'releases.no_artists': 'Du folgst noch keinem Künstler.',
  'releases.attribution': 'Daten: MusicBrainz (CC0), Cover: Cover Art Archive.',
};
