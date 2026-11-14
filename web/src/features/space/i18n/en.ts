/** English UI texts of the space feature (#64). */
export const en = {
  'widget.space.title': 'Space',
  'widget.space.description':
    'Next rocket launches (Launch Library 2) and NASA’s Astronomy Picture of the Day.',
  'settings.section.space': 'Space',
  'space.intro':
    'Both sources are off until you switch them on; afterwards they are only loaded while a widget shows them.',
  'space.launches_on': 'Show rocket launches (Launch Library 2)',
  'space.picture_on': 'Show the Astronomy Picture of the Day (NASA)',
  'space.nasa_key': 'NASA API key (optional)',
  'space.nasa_key_stored': 'stored – leave empty to keep',
  'space.nasa_key_hint':
    'Without your own free key from api.nasa.gov the shared DEMO_KEY is used, which has a low limit per IP address.',
  'space.background': 'Picture of the day as background',
  'space.background_hint':
    'Use this address as tab background (Edit → 🖼) or under Appearance. It always shows today’s picture, loaded by this server.',
  'space.background_copy': 'Copy background address',
  'space.saved': 'Saved.',
  'space.show': 'Show',
  'space.show_both': 'Launches and picture',
  'space.show_launches': 'Launches only',
  'space.show_picture': 'Picture only',
  'space.count': 'Launches: {n}',
  'space.switched_off': 'Switch the space sources on under Settings → Space.',
  'space.next_launches': 'Next launches',
  'space.no_launches': 'No launches announced.',
  'space.now': 'now',
  'space.in_minutes': 'in {n} min',
  'space.in_hours': 'in {n} h',
  'space.in_days': 'in {n} days',
  'space.live': 'Live',
  'space.status_go': 'Go',
  'space.status_tbd': 'TBD',
  'space.status_tbc': 'TBC',
  'space.status_hold': 'Hold',
  'space.status_in_flight': 'In flight',
  'space.status_success': 'Success',
  'space.status_failure': 'Failure',
  'space.status_partial_failure': 'Partial failure',
  'space.status_unknown': '?',
  'space.picture': 'Picture of the day',
  'space.video_today': 'Today’s picture is a video.',
  'space.open': 'Open',
  'space.more': 'More',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
