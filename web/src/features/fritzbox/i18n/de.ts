import type { en } from './en';

/** German UI texts of the fritzbox feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.fritzbox_calls.title': 'FRITZ!Box-Anrufe',
  'widget.fritzbox_calls.description':
    'Letzte Anrufe der FRITZ!Box: alle, verpasste, eingehende oder ausgehende.',
  'fritzbox.show_uptime': 'Verbindungsdauer anzeigen',
  'fritzbox.show_ip': 'Externe IP-Adresse anzeigen',
  'fritzbox.show_rates': 'Aktuellen Down- und Upload anzeigen',
  'fritzbox.calls_filter': 'Anrufe',
  'fritzbox.filter.all': 'Alle Anrufe',
  'fritzbox.filter.missed': 'Verpasste',
  'fritzbox.filter.incoming': 'Eingehende',
  'fritzbox.filter.outgoing': 'Ausgehende',
  'fritzbox.calls_count': 'Anzahl der Anrufe (5–50)',
  'fritzbox.show_numbers': 'Nummern von Anrufenden ohne Namen anzeigen',
  'fritzbox.save': 'Speichern',
  'fritzbox.no_calls': 'Keine Anrufe.',
  'fritzbox.call.incoming': 'eingehend',
  'fritzbox.call.missed': 'verpasst',
  'fritzbox.call.outgoing': 'ausgehend',
  'fritzbox.call.rejected': 'abgewiesen',
  'fritzbox.call.active': 'aktiv',
};
