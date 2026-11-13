/** English UI texts of the media feature (#193). */
export const en = {
  'widget.media.title': 'Media server',
  'widget.media.description': 'Now playing and recently added on Jellyfin or Plex.',
  'settings.section.media': 'Media server',
  'media.connect_under_settings': 'Connect Jellyfin or Plex under Settings → Media server.',
  'media.now_playing': 'Now playing',
  'media.nothing_playing': 'Nothing is playing.',
  'media.recently_added': 'Recently added',
  'media.server': 'Media server',
  'media.url': '{name} URL',
  'media.jellyfin_key': 'API key {value}',
  'media.plex_token': 'Plex token {value}',
  'media.jellyfin_hint':
    'Create an API key in Jellyfin under Dashboard → API keys. It is stored encrypted; posters are loaded through trisse-home.',
  'media.plex_hint':
    'Use the X-Plex-Token of your account (Plex support article "Finding an authentication token"). It is stored encrypted; posters are loaded through trisse-home.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
