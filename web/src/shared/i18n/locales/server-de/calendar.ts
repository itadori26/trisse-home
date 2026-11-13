import type { ServerMessageKey } from '../../../api-client/generated/server-messages/calendar';

/**
 * German translations of the backend error messages of `calendar` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most 10 calendars are supported.
  srv_326bb72e415244d3: 'Es werden höchstens 10 Kalender unterstützt.',
  // Events can be at most one year long.
  srv_38ba62864162fcec: 'Termine können höchstens ein Jahr lang sein.',
  // No calendar is configured yet.
  srv_92fe9be68e4fbeb4: 'Es ist noch kein Kalender eingerichtet.',
  // Please choose one of the configured calendars.
  srv_f5627673a11b56bd: 'Bitte wähle einen der eingerichteten Kalender.',
  // Please enter a valid start and end.
  srv_b71f32ebd27a0b6a: 'Bitte gib einen gültigen Beginn und ein gültiges Ende ein.',
  // Please enter at least one calendar URL.
  srv_a4d2757d3ca7d004: 'Bitte mindestens eine Kalender-URL eingeben.',
  // The calendar server is not reachable.
  srv_3e73df212f6df359: 'Der Kalenderserver ist nicht erreichbar.',
  // The event must end after it starts.
  srv_db999c150db53756: 'Der Termin muss nach seinem Beginn enden.',
  // The location must be at most 200 characters long.
  srv_eeec5bbc9fb36341: 'Der Ort darf höchstens 200 Zeichen lang sein.',
  // The period must be positive and at most 42 days long.
  srv_066ea148faee11ae: 'Der Zeitraum muss positiv und höchstens 42 Tage lang sein.',
  // The title must be between 1 and 200 characters long.
  srv_f79c303515b1250c: 'Der Titel muss zwischen 1 und 200 Zeichen lang sein.',
  // from and to must be RFC 3339 timestamps.
  srv_ac92ff8f414a83b9: 'from und to müssen RFC-3339-Zeitstempel sein.',
  // At most {max} calendar subscriptions are supported.
  srv_f74740ac20c25101: 'Höchstens {max} Kalender-Abos werden unterstützt.',
  // Could not load the ICS calendar: {detail}
  srv_97f22fc39ea1794a: 'Der ICS-Kalender konnte nicht geladen werden: {detail}',
  // The calendar subscription does not exist.
  srv_210a34361dc2d298: 'Das Kalender-Abo existiert nicht.',
  // The colour must look like #3b6ef5.
  srv_5fd12f163f5132ff: 'Die Farbe muss wie #3b6ef5 aussehen.',
  // The refresh interval must be between 15 minutes and 24 hours.
  srv_99e5a87eee94b355: 'Das Aktualisierungsintervall muss zwischen 15 Minuten und 24 Stunden liegen.',
};
