import type { en } from './en';

/** German UI texts of the drop zone; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.drop.title': 'Ablage',
  'widget.drop.description':
    'Text, Links und kleine Dateien zwischen deinen Geräten austauschen; Einträge laufen automatisch ab.',
  'drop.zone': 'Ablage: Text einfügen oder Dateien hierher ziehen',
  'drop.text': 'Text oder Link',
  'drop.placeholder': 'Text oder Link einfügen oder Dateien hierher ziehen …',
  'drop.expiry': 'Aufbewahren für',
  'drop.expiry_1': '1 Stunde',
  'drop.expiry_24': '24 Stunden',
  'drop.expiry_168': '7 Tage',
  'drop.file': 'Datei …',
  'drop.drop': 'Ablegen',
  'drop.copy': 'Kopieren',
  'drop.download': '{name} herunterladen',
  'drop.delete': 'Löschen',
  'drop.empty': 'Noch leer. Leg etwas ab, um es auf einem anderen Gerät abzuholen.',
  'drop.expires': 'läuft {when} ab',
  'drop.in_minutes': 'in {n} Min.',
  'drop.in_hours': 'in {n} Std.',
  'drop.in_days': 'in {n} Tagen',
  'drop.file_empty': 'Die Datei ist leer.',
  'drop.file_too_large': 'Die Datei darf höchstens 25 MB groß sein.',
};
