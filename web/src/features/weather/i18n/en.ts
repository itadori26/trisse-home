/** English UI texts added to the weather feature (#118, #193). */
export const en = {
  'weather.wind_from': 'Wind from {direction}',
  'weather.wind_n': 'N',
  'weather.wind_ne': 'NE',
  'weather.wind_e': 'E',
  'weather.wind_se': 'SE',
  'weather.wind_s': 'S',
  'weather.wind_sw': 'SW',
  'weather.wind_w': 'W',
  'weather.wind_nw': 'NW',
  'weather.wind_name_n': 'the north',
  'weather.wind_name_ne': 'the north-east',
  'weather.wind_name_e': 'the east',
  'weather.wind_name_se': 'the south-east',
  'weather.wind_name_s': 'the south',
  'weather.wind_name_sw': 'the south-west',
  'weather.wind_name_w': 'the west',
  'weather.wind_name_nw': 'the north-west',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
