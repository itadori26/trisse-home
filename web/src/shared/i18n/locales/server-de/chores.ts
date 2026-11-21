import type { ServerMessageKey } from '../../../api-client/generated/server-messages/chores';

/** German translations of the backend messages of `chores` (#102, #683). */
export const serverDe: Record<ServerMessageKey, string> = {
  // 1 task today
  srv_b02f4545b938ddfc: 'Heute 1 Aufgabe',
  // A chore can rotate between at most 8 people.
  srv_65689b528e357fa1: 'Eine Aufgabe kann höchstens zwischen 8 Personen wechseln.',
  // A name may be at most 40 characters long.
  srv_c582a0f40016b0a3: 'Ein Name darf höchstens 40 Zeichen lang sein.',
  // No chores yet
  srv_8fb7d91db33f9672: 'Noch keine Aufgaben',
  // Nothing to do today
  srv_53bf4bebb859bdf9: 'Heute ist nichts zu tun',
  // Repeat a chore every 1 to 365 days.
  srv_86c8345f05ad8b7e: 'Eine Aufgabe wiederholt sich alle 1 bis 365 Tage.',
  // The chore does not exist.
  srv_2dd4860e77e0ca9e: 'Die Aufgabe existiert nicht.',
  // The chore must be between 1 and 80 characters long.
  srv_1bb925b981c6d045: 'Die Aufgabe muss zwischen 1 und 80 Zeichen lang sein.',
  // {count} tasks today
  srv_4548096ca54179fb: 'Heute {count} Aufgaben',
  // {title} ({name})
  srv_bd31a1ae5c8ba411: '{title} ({name})',
};
