import type { en } from './en';

/** German UI texts of the device feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'settings.section.device': 'Gerät',
  'device.question': 'Wie möchtest du dieses Gerät nutzen?',
  'device.headless': 'Nur Server',
  'device.headless_hint': 'trisse-home auf Handy, Tablet oder Computer öffnen.',
  'device.kiosk': 'Server + Anzeige',
  'device.kiosk_hint':
    'Zusätzlich das Dashboard im Vollbild auf dem am Pi angeschlossenen Bildschirm zeigen (installiert einen schlanken Browser, mehrere hundert MB).',
  'device.display_connected': 'An dieses Gerät ist ein Bildschirm angeschlossen.',
  'device.display_disconnected': 'Gerade ist kein Bildschirm angeschlossen.',
  'device.display_unknown': 'Ob ein Bildschirm angeschlossen ist, ist unbekannt.',
  'device.apply': 'Übernehmen',
  'device.pending': 'Angefordert – das Gerät übernimmt es gleich …',
  'device.installing': 'Das Gerät richtet die Anzeige ein, das kann ein paar Minuten dauern …',
  'device.failed': 'Die Änderung ist fehlgeschlagen: {error}',
  'device.active_kiosk': 'Das Dashboard wird auf dem angeschlossenen Bildschirm gezeigt.',
  'device.active_headless': 'Dieses Gerät läuft nur als Server.',
  'device.later': 'Du kannst das jederzeit unter Einstellungen → Gerät ändern.',
  'device.settings_hint':
    'Wähle, ob dieser Raspberry Pi das Dashboard auch auf seinem eigenen Bildschirm zeigt.',
};
