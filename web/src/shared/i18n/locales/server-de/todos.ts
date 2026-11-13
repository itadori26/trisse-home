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
};
