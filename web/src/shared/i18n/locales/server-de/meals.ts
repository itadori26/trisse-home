import type { ServerMessageKey } from '../../../api-client/generated/server-messages/meals';

/** German translations of the backend messages of `meals` (#102, #682). */
export const serverDe: Record<ServerMessageKey, string> = {
  // A dish may have at most 40 ingredients.
  srv_0a92f9012467690f: 'Ein Gericht darf höchstens 40 Zutaten haben.',
  // An ingredient may be at most 100 characters long.
  srv_3691a82b79866d12: 'Eine Zutat darf höchstens 100 Zeichen lang sein.',
  // Ask for 1 to 42 days.
  srv_ec7d3b103e184fbc: 'Bitte 1 bis 42 Tage abfragen.',
  // Nothing planned for today
  srv_d88ce767c858967e: 'Für heute ist nichts geplant',
  // The dish does not exist.
  srv_d2f70f839ca60227: 'Das Gericht existiert nicht.',
  // The dish must be between 1 and 120 characters long.
  srv_cd8bf9f8a954422f: 'Das Gericht muss zwischen 1 und 120 Zeichen lang sein.',
  // The link must be a web address starting with http:// or https://.
  srv_b429d991f10042c7: 'Der Link muss eine Webadresse sein, die mit http:// oder https:// beginnt.',
  // The meal must be breakfast, lunch, dinner or empty.
  srv_6d9cc506ee0c6722: 'Die Mahlzeit muss Frühstück, Mittagessen, Abendessen oder leer sein.',
  // Today: {dish}
  srv_9482b4086fdf3088: 'Heute: {dish}',
  // and {count} more
  srv_bdb37f2bc1177fca: 'und {count} weitere',
};
