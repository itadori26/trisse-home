/** English UI texts of the arr feature (#193). */
export const en = {
  'widget.arr.title': 'Sonarr / Radarr',
  'widget.arr.description':
    'Upcoming episodes and movies, the download queue and health warnings of Sonarr, Radarr and Prowlarr.',
  'settings.section.arr': 'Sonarr / Radarr',
  'arr.connect_under_settings': 'Add Sonarr, Radarr or Prowlarr under Settings → Sonarr / Radarr.',
  'arr.queue': 'Downloads',
  'arr.queue_empty': 'No downloads.',
  'arr.upcoming': 'Upcoming',
  'arr.nothing_upcoming': 'Nothing in the next two weeks.',
  'arr.health': 'Health warnings',
  'arr.missing': '{name}: {count} missing',
  'arr.time_left': '{value} left',
  'arr.unreachable': 'Not reachable: {detail}',
  'arr.app': 'App',
  'arr.name': 'Name',
  'arr.api_key': 'API key {value}',
  'arr.key_hint':
    'The API key is under Settings → General in each app. It is stored encrypted and never shown again.',
  'arr.add': 'Add instance',
  'arr.edit': 'Edit {name}',
  'arr.delete': 'Delete {name}',
  'arr.invalid_url': 'The URL must start with http:// or https://.',
  'arr.key_required': 'Please enter the API key.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
