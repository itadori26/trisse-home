/** English UI texts of the gamedeals feature (#193). */
export const en = {
  'widget.gamedeals.title': 'Game deals',
  'widget.gamedeals.description': 'Discounts on your Steam wishlist and free games on the Epic Games Store.',
  'settings.section.gamedeals': 'Game deals',
  'gamedeals.connect_under_settings': 'Set up game deals under Settings → Game deals.',
  'gamedeals.free': 'Free on Epic',
  'gamedeals.until': 'until {date}',
  'gamedeals.from': 'from {date}',
  'gamedeals.wishlist': 'Steam wishlist on sale',
  'gamedeals.nothing': 'No deals right now.',
  'gamedeals.unreachable': '{source} could not be loaded.',
  'gamedeals.steam_id': 'Steam profile URL or ID (optional)',
  'gamedeals.steam_hint':
    'The wishlist must be public (Steam → Edit profile → Privacy settings → Game details: Public).',
  'gamedeals.country': 'Country',
  'gamedeals.min_discount': 'Minimum discount in %',
  'gamedeals.epic': 'Show free games of the Epic Games Store',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
