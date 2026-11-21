/** English UI texts of the places feature (#591). */
export const en = {
  'settings.section.connections': 'Places',
  'places.title': 'Places',
  'places.intro':
    'Enter a place once and choose it in every widget that needs a location: weather, rain radar, fuel prices and nearby stops.',
  'places.empty': 'No places saved yet.',
  'places.used_by': 'Used by {list}',
  'places.unused': 'Not used by any widget',
  'places.rename': 'Rename',
  'places.rename_place': 'Rename {name}',
  'places.remove': 'Remove',
  'places.remove_place': 'Remove {name}',
  'places.confirm_remove': 'Remove “{name}”?',
  'places.confirm_remove_used':
    '“{name}” is used by {list}. These widgets keep their location but no longer follow the place. Remove it?',
  'places.add': 'Add a place',
  'places.search': 'Search',
  'places.search_label': 'Search an address or town',
  'places.search_placeholder': 'e.g. Oldenburg',
  'places.no_results': 'Nothing found.',
  'places.my_location': 'Use my location',
  'places.locating': 'Finding your location…',
  'places.location_denied': 'The browser did not share the location.',
  'places.my_location_name': 'My location',
  'places.name': 'Name',
  'places.save': 'Save place',
  'places.cancel': 'Cancel',
  'places.name_empty': 'Give the place a name.',
  'places.name_too_long': 'The name can have at most 80 characters.',
  'places.coordinates': '{lat}, {lon}',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
