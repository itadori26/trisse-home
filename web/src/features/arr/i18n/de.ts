import type { en } from './en';

/** German UI texts of the arr feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.arr.title': 'Sonarr / Radarr',
  'widget.arr.description':
    'Kommende Folgen und Filme, Download-Warteschlange und Warnungen von Sonarr, Radarr und Prowlarr.',
  'settings.section.arr': 'Sonarr / Radarr',
  'arr.connect_under_settings':
    'Füge Sonarr, Radarr oder Prowlarr unter Einstellungen → Sonarr / Radarr hinzu.',
  'arr.queue': 'Downloads',
  'arr.queue_empty': 'Keine Downloads.',
  'arr.upcoming': 'Demnächst',
  'arr.nothing_upcoming': 'Nichts in den nächsten zwei Wochen.',
  'arr.health': 'Warnungen',
  'arr.missing': '{name}: {count} fehlen',
  'arr.time_left': 'noch {value}',
  'arr.unreachable': 'Nicht erreichbar: {detail}',
  'arr.app': 'App',
  'arr.name': 'Name',
  'arr.api_key': 'API-Schlüssel {value}',
  'arr.key_hint':
    'Den API-Schlüssel findest du in jeder App unter Settings → General. Er wird verschlüsselt gespeichert und nie wieder angezeigt.',
  'arr.add': 'Instanz hinzufügen',
  'arr.edit': '{name} bearbeiten',
  'arr.delete': '{name} löschen',
  'arr.invalid_url': 'Die URL muss mit http:// oder https:// beginnen.',
  'arr.key_required': 'Bitte gib den API-Schlüssel ein.',
};
