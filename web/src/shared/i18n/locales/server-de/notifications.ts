import type { ServerMessageKey } from '../../../api-client/generated/server-messages/notifications';

/**
 * German translations of the backend error messages of `notifications` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A maintenance window must end after it starts and last at most 7 days.
  srv_a912b3e2ea2b8170: 'Ein Wartungsfenster muss nach seinem Beginn enden und darf höchstens 7 Tage dauern.',
  // Choose what the maintenance applies to.
  srv_4e9dc8e29fa35fbf: 'Bitte wählen, wofür die Wartung gilt.',
  // Invalid scope {scope}: use service:<name>, container:<name> or *.
  srv_d6e30a440a4e7b86: 'Ungültige Angabe {scope}: bitte service:<Name>, container:<Name> oder * verwenden.',
  // Please enter a valid sender address.
  srv_ac8a8ae2b384e281: 'Bitte eine gültige Absenderadresse eingeben.',
  // Please enter and save a topic first.
  srv_dd3a700b0be13baf: 'Bitte zuerst ein Topic eingeben und speichern.',
  // Please enter the SMTP server name.
  srv_5057fd283cbaf8f6: 'Bitte den Namen des SMTP-Servers eingeben.',
  // Please enter valid recipient addresses, separated by commas.
  srv_e9295c1002fdd2ba: 'Bitte gültige Empfängeradressen eingeben, durch Kommas getrennt.',
  // Quiet hours need a start and an end such as 22:00 and 07:00.
  srv_04eab3a4b4f8b837: 'Ruhezeiten brauchen einen Beginn und ein Ende, z. B. 22:00 und 07:00.',
  // TLS must be starttls, tls or none.
  srv_8ba7cc04f14ed353: 'TLS muss starttls, tls oder none sein.',
  // The certificate warning must be between 1 and 90 days.
  srv_55663fc836edb6a7: 'Die Zertifikatswarnung muss zwischen 1 und 90 Tagen liegen.',
  // The channel does not exist.
  srv_e97962ce6df48c04: 'Der Kanal existiert nicht.',
  // The channel type must be gotify, webhook or email.
  srv_369ee646cdd39767: 'Die Kanalart muss gotify, webhook oder email sein.',
  // The disk threshold must be between 50 and 99 %.
  srv_1c6e2427a40c8139: 'Der Speicher-Schwellwert muss zwischen 50 und 99 % liegen.',
  // The maintenance window does not exist.
  srv_d6a9caae13404f6e: 'Das Wartungsfenster existiert nicht.',
  // The port must be between 1 and 65535.
  srv_6566f488a34e8fc3: 'Der Port muss zwischen 1 und 65535 liegen.',
  // The template must be at most 4000 characters long.
  srv_83146ba690ec10f4: 'Die Vorlage darf höchstens 4000 Zeichen lang sein.',
  // The topic may contain letters, digits, - and _ (max. 64 characters).
  srv_218a25cf9751b7dd: 'Das Topic darf Buchstaben, Ziffern, - und _ enthalten (max. 64 Zeichen).',
  // There are already 10 channels.
  srv_53e6d03eb73e5456: 'Es gibt bereits 10 Kanäle.',
  // Unknown alert type {type}.
  srv_fc5b4d698d5d3c6f: 'Unbekannte Benachrichtigungsart {type}.',
};
