/** English UI texts of the smarthome feature (#193). */
export const en = {
  'widget.smarthome.title': 'Smart home',
  'widget.smarthome.description':
    'States of Home Assistant devices and sensors; optional switches and buttons.',
  'settings.section.smarthome': 'Home Assistant',
  'smarthome.connect_under_settings': 'Connect Home Assistant under Settings → Home Assistant.',
  'smarthome.choose_entities': 'Choose entities in the widget settings.',
  'smarthome.unavailable': 'unavailable',
  'smarthome.on': 'on',
  'smarthome.off': 'off',
  'smarthome.turn_on': 'Turn on',
  'smarthome.turn_off': 'Turn off',
  'smarthome.press': 'Press',
  'smarthome.confirm_toggle': 'Switch {name}?',
  'smarthome.confirm_press': 'Press {name}?',
  'smarthome.search_entities': 'Search entities',
  'smarthome.selected': '{count} of at most {max} selected',
  'smarthome.url': 'Home Assistant URL',
  'smarthome.token': 'Long-lived access token {value}',
  'smarthome.token_hint':
    'Create the token in Home Assistant under your profile → Security → Long-lived access tokens. It is stored encrypted and never shown again.',
  'smarthome.actions_hint':
    'Switches and buttons stay read-only unless SMARTHOME_ACTIONS_ENTITIES lists their entity IDs.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
