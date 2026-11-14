import type { ServerMessageKey } from '../../../api-client/generated/server-messages/parcels';

/**
 * German translations of the backend error messages of `parcels` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most 30 parcels can be tracked.
  srv_5343c52cdf9d2afd: 'Es können höchstens 30 Sendungen verfolgt werden.',
  // DHL does not know this tracking number (yet).
  srv_24f489be60f9182b: 'DHL kennt diese Sendungsnummer (noch) nicht.',
  // DHL rejected the API key.
  srv_099fe605093d8b87: 'DHL hat den API-Schlüssel abgelehnt.',
  // Delivered parcels can be kept for 0 to 30 days.
  srv_7cb71b2e113119ec: 'Zugestellte Sendungen können 0 bis 30 Tage angezeigt werden.',
  // Please add your DHL API key in the settings first.
  srv_7d0f681f466f8e83: 'Bitte zuerst deinen DHL-API-Schlüssel in den Einstellungen hinterlegen.',
  // The daily DHL limit is reached; it will be retried later.
  srv_2d07f911dc00e3dc: 'Das tägliche DHL-Limit ist erreicht; es wird später erneut versucht.',
  // The label must be at most 60 characters long.
  srv_64e5ff21727bf802: 'Die Bezeichnung darf höchstens 60 Zeichen lang sein.',
  // The parcel does not exist.
  srv_f35b41c22f1a21c2: 'Die Sendung existiert nicht.',
  // This carrier is not supported.
  srv_06ef324c0db01986: 'Dieser Zusteller wird nicht unterstützt.',
  // Tracking numbers consist of 8 to 35 letters and digits.
  srv_e4cd1895823b2401: 'Sendungsnummern bestehen aus 8 bis 35 Buchstaben und Ziffern.',
  // Check the mailbox every 5 minutes to 24 hours.
  srv_cee130f3904d48b7: 'Prüfe das Postfach alle 5 Minuten bis 24 Stunden.',
  // Enter the IMAP server name, e.g. imap.example.com.
  srv_91809f90e8dfc4fb: 'Gib den Namen des IMAP-Servers ein, z. B. imap.example.com.',
  // The port must be between 1 and 65535 (IMAP over TLS: 993).
  srv_92285ba58a255b47: 'Der Port muss zwischen 1 und 65535 liegen (IMAP über TLS: 993).',
  // User name and app password are required.
  srv_970344b28687ee67: 'Benutzername und App-Passwort sind erforderlich.',
  // User name, password and folder must not contain line breaks.
  srv_b08dda17b8ef851b: 'Benutzername, Passwort und Ordner dürfen keine Zeilenumbrüche enthalten.',
};
