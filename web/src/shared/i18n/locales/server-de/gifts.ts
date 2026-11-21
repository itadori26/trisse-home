import type { ServerMessageKey } from '../../../api-client/generated/server-messages/gifts';

/** German translations of the backend messages of `gifts` (#102, #684). */
export const serverDe: Record<ServerMessageKey, string> = {
  // 1 gift idea
  srv_68c8d9fe8ab372f1: '1 Geschenkidee',
  // No open gift ideas
  srv_2a36cc5a6ff84668: 'Keine offenen Geschenkideen',
  // The gift idea does not exist.
  srv_5e0a58598f59dbce: 'Die Geschenkidee existiert nicht.',
  // The idea must be between 1 and 120 characters long.
  srv_cc83d8504734d7c2: 'Die Idee muss zwischen 1 und 120 Zeichen lang sein.',
  // The link must be a web address starting with http:// or https://.
  srv_b429d991f10042c7: 'Der Link muss eine Webadresse sein, die mit http:// oder https:// beginnt.',
  // The note may be at most 500 characters long.
  srv_279687ba82800dcf: 'Die Notiz darf höchstens 500 Zeichen lang sein.',
  // {count} gift ideas
  srv_8bc0d4ea1b74675e: '{count} Geschenkideen',
};
