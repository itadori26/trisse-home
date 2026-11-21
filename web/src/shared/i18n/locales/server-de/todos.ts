import type { ServerMessageKey } from '../../../api-client/generated/server-messages/todos';

/**
 * German translations of the backend error messages of `todos` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A reminder needs a due date.
  srv_fc8a2589829c3428: 'Eine Erinnerung braucht ein Fälligkeitsdatum.',
  // Reminders need a time like 08:00 and at most 30 days before the due date.
  srv_33507329e3cf1562:
    'Erinnerungen brauchen eine Uhrzeit wie 08:00 und höchstens 30 Tage vor dem Fälligkeitsdatum.',
  // Repeat every 1–365 days, weeks or months; weekly repeats need at least one weekday.
  srv_5f173b47eee5e70d:
    'Wiederholung alle 1–365 Tage, Wochen oder Monate; wöchentliche Wiederholungen brauchen mindestens einen Wochentag.',
  // The due date must look like 2026-12-24.
  srv_8fd5a1f81a5668ce: 'Das Fälligkeitsdatum muss wie 2026-12-24 aussehen.',
  // The to-do does not exist.
  srv_39e64da309876466: 'Das To-do existiert nicht.',
  // The to-do must have between 1 and 300 characters.
  srv_deed27849449d0b5: 'Das To-do muss zwischen 1 und 300 Zeichen haben.',
  // The to-do was changed elsewhere in the meantime.
  srv_b4f6dd576f1534c2: 'Die Aufgabe wurde inzwischen an anderer Stelle geändert.',
  // At most 20 to-do lists can be connected.
  srv_a284ac8e205f55f5: 'Es können höchstens 20 To-do-Listen verbunden werden.',
  // The name must have between 1 and 80 characters.
  srv_59575ee33ddb1b96: 'Der Name muss zwischen 1 und 80 Zeichen haben.',
  // The to-do list does not exist (any more).
  srv_984c25c4aa086942: 'Die To-do-Liste existiert nicht (mehr).',
  // This list does not support repeating to-dos or reminders.
  srv_b8a4d7c2f106adb7: 'Diese Liste unterstützt keine wiederholten To-dos und keine Erinnerungen.',
  // Unknown kind of to-do list.
  srv_e28b70dde38477c1: 'Unbekannte Art von To-do-Liste.',
  // Choose a Todoist project.
  srv_e21c32bf57ab4be8: 'Wähle ein Todoist-Projekt.',
  // Enter the API token from Todoist's settings (Integrations → Developer).
  srv_844dc15540ba01b8: 'Gib das API-Token aus den Todoist-Einstellungen ein (Integrationen → Entwickler).',
  // The service asks to slow down. The list is loaded again in a minute.
  srv_31173fc19f6e7f41: 'Der Dienst bittet um weniger Anfragen. Die Liste wird in einer Minute neu geladen.',
  // Choose a list.
  srv_01e6a58b40c54321: 'Wähle eine Liste.',
  // The sign-in did not succeed. Check client ID, client secret and the registered redirect address.
  srv_f61fe037ae3068a8:
    'Die Anmeldung hat nicht geklappt. Prüfe Client-ID, Client-Secret und die registrierte Weiterleitungsadresse.',
  // The sign-in has expired or was revoked. Disconnect the list and connect it again.
  srv_77c57fca0b0dafe3:
    'Die Anmeldung ist abgelaufen oder wurde widerrufen. Trenne die Liste und verbinde sie neu.',
  // Sharing is not available here.
  srv_821a541a7d6c7b9f: 'Teilen ist hier nicht verfügbar.',
  // Nothing to do
  srv_8474d36c18713018: 'Nichts zu tun',
  // {count} open
  srv_fe19a87a4c31a1ac: '{count} offen',
  // {count} overdue
  srv_4fd610086ba2117c: '{count} überfällig',
};
