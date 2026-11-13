import type { ServerMessageKey } from '../../../api-client/generated/server-messages/services';

/**
 * German translations of the backend error messages of `services` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // No MAC address is stored for this service.
  srv_591001f7788fb502: 'Für diesen Dienst ist keine MAC-Adresse hinterlegt.',
  // The check URL must start with http:// or https://.
  srv_49713b674a89dd04: 'Die Prüf-URL muss mit http:// oder https:// beginnen.',
  // The description must be at most 200 characters long.
  srv_b6d4e2000983a0fe: 'Die Beschreibung darf höchstens 200 Zeichen lang sein.',
  // The icon is too long.
  srv_42e741b2bbfc2695: 'Das Icon ist zu lang.',
  // The new order must contain every service exactly once.
  srv_a75b491a5e89aabe: 'Die neue Reihenfolge muss jeden Dienst genau einmal enthalten.',
  // The service does not exist.
  srv_df6166e3647fbc04: 'Der Dienst existiert nicht.',
  // Docker discovery needs the docker module to be enabled.
  srv_1f3976061912504c: 'Die Docker-Autodiscovery braucht das aktivierte Docker-Modul.',
  // The interval must be between 1 and {max} minutes.
  srv_c2aa1fabd915da16: 'Das Intervall muss zwischen 1 und {max} Minuten liegen.',
  // Unknown merge rule.
  srv_c64f8cf81797f821: 'Unbekannte Abgleichsregel.',
};
