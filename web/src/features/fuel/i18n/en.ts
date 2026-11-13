/** English UI texts of the fuel feature (#193). */
export const en = {
  'widget.fuel.title': 'Fuel prices',
  'widget.fuel.description':
    'Cheapest stations around you or your favourite stations (Germany, Tankerkönig).',
  'settings.section.fuel': 'Fuel prices',
  'fuel.connect_under_settings': 'Set up fuel prices under Settings → Fuel prices.',
  'fuel.closed': 'closed',
  'fuel.favourite': 'Favourite {name}',
  'fuel.no_favourites': 'Mark stations with ☆ to see them here.',
  'fuel.no_stations': 'No stations in the radius.',
  'fuel.attribution': 'Data: MTS-K via Tankerkönig (CC BY 4.0)',
  'fuel.view': 'Show',
  'fuel.view_cheapest': 'Cheapest stations',
  'fuel.view_favourites': 'Favourite stations',
  'fuel.search_place': 'Search a place',
  'fuel.location': 'Location: {place} ({lat}, {lon})',
  'fuel.choose_location': 'Please choose a location.',
  'fuel.lat': 'Latitude',
  'fuel.lon': 'Longitude',
  'fuel.fuel': 'Fuel',
  'fuel.radius': 'Radius in km',
  'fuel.api_key': 'Tankerkönig API key {value}',
  'fuel.key_hint':
    'Request a free API key at creativecommons.tankerkoenig.de. Prices are loaded at most every five minutes and only while the widget is open.',
  'fuel.invalid_radius': 'The radius must be between 1 and 25 km.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
