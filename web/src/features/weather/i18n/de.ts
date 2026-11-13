import type { en } from './en';

/** German UI texts added to the weather feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'weather.wind_from': 'Wind aus {direction}',
  'weather.wind_n': 'N',
  'weather.wind_ne': 'NO',
  'weather.wind_e': 'O',
  'weather.wind_se': 'SO',
  'weather.wind_s': 'S',
  'weather.wind_sw': 'SW',
  'weather.wind_w': 'W',
  'weather.wind_nw': 'NW',
  'weather.wind_name_n': 'Norden',
  'weather.wind_name_ne': 'Nordosten',
  'weather.wind_name_e': 'Osten',
  'weather.wind_name_se': 'Südosten',
  'weather.wind_name_s': 'Süden',
  'weather.wind_name_sw': 'Südwesten',
  'weather.wind_name_w': 'Westen',
  'weather.wind_name_nw': 'Nordwesten',
};
