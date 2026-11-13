import type { ServerMessageKey } from '../../../api-client/generated/server-messages/countdown';

/**
 * German translations of the backend error messages of `countdown` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Enter a date like 2026-12-24.
  srv_22fc413ef29251fa: 'Bitte ein Datum wie 2026-12-24 eingeben.',
  // graceDays must be a number.
  srv_1bef734e6f2cdc94: 'graceDays muss eine Zahl sein.',
  // The countdown does not exist.
  srv_1532da22068d27c4: 'Den Countdown gibt es nicht.',
  // The title must be between 1 and 120 characters long.
  srv_28a208881a6ef97d: 'Der Titel muss zwischen 1 und 120 Zeichen lang sein.',
  // Use a single emoji.
  srv_61031ab185b65e33: 'Bitte nur ein einzelnes Emoji verwenden.',
};
