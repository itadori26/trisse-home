import type { ServerMessageKey } from '../../../api-client/generated/server-messages/waste';

/**
 * German translations of the backend error messages of `waste` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most 20 rules are supported.
  srv_741879779a784a6e: 'Es werden höchstens 20 Regeln unterstützt.',
  // Colors must look like #1a2b3c.
  srv_100561a937e88349: 'Farben müssen wie #1a2b3c aussehen.',
  // Could not load the waste calendar: {detail}
  srv_57ceb20ccebc67bb: 'Der Abfallkalender konnte nicht geladen werden: {detail}',
  // Every rule needs a match text of at most 40 characters.
  srv_8e50b118b5e6b205: 'Jede Regel braucht einen Suchtext mit höchstens 40 Zeichen.',
  // Icons must be at most 8 characters long.
  srv_72f2ef28f00d9995: 'Icons dürfen höchstens 8 Zeichen lang sein.',
  // Labels must be at most 40 characters long.
  srv_4cd4b235493d7de0: 'Bezeichnungen dürfen höchstens 40 Zeichen lang sein.',
  // No waste calendar is configured yet.
  srv_d1f1feb231209e5a: 'Es ist noch kein Abfallkalender eingerichtet.',
};
