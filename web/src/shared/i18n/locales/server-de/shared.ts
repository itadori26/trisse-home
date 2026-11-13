import type { ServerMessageKey } from '../../../api-client/generated/server-messages/shared';

/**
 * German translations of the backend error messages of `shared` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Connections to this address are not allowed.
  srv_607f6c84226cd658: 'Verbindungen zu dieser Adresse sind nicht erlaubt.',
  // Enter a time like 07:00.
  srv_45cc329a438ca49f: 'Bitte eine Uhrzeit wie 07:00 eingeben.',
  // Enter an http(s) address with a host, e.g. https://nas.local:8443.
  srv_c551eccad97486f8: 'Bitte eine http(s)-Adresse mit Host eingeben, z. B. https://nas.local:8443.',
  // Enter only the server address, without a path.
  srv_871c0989570954df: 'Bitte nur die Serveradresse ohne Pfad eingeben.',
  // Link-local and cloud metadata addresses are not allowed.
  srv_62a7030824310857: 'Link-lokale Adressen und Cloud-Metadaten-Adressen sind nicht erlaubt.',
  // Remove the user name and password from the address and use the fields below.
  srv_83cc78074d3e2ff6:
    'Bitte Benutzername und Passwort aus der Adresse entfernen und die Felder darunter nutzen.',
  // Storing credentials is disabled in the demo.
  srv_ea4ce39d77587136: 'Das Speichern von Zugangsdaten ist in der Demo deaktiviert.',
  // The address must not contain a query (?…) or fragment (#…).
  srv_693550da2f4a6610: 'Die Adresse darf keine Abfrage (?…) und keinen Anker (#…) enthalten.',
  // The address or account changed: enter the password or token again.
  srv_4210fb1d03642aa8: 'Adresse oder Konto wurden geändert: Bitte Passwort oder Token erneut eingeben.',
  // The certificate of the server is not trusted.
  srv_73afeaf5a99f225a: 'Dem Zertifikat des Servers wird nicht vertraut.',
  // The MAC address must look like aa:bb:cc:dd:ee:ff.
  srv_022f9a869dd885c4: 'Die MAC-Adresse muss so aussehen: aa:bb:cc:dd:ee:ff.',
  // The server answered, but the path was not found. Check the address.
  srv_0d59baf576015864: 'Der Server antwortet, aber der Pfad wurde nicht gefunden. Bitte die Adresse prüfen.',
  // The server could not be reached. Check the address and port.
  srv_71b0db214ed1ca3d: 'Der Server ist nicht erreichbar. Bitte Adresse und Port prüfen.',
  // The server did not answer in time.
  srv_5cd46d3ed824e037: 'Der Server hat nicht rechtzeitig geantwortet.',
  // The server rejected the login. Check user name, password or token.
  srv_d6ecea29338eeff4:
    'Der Server hat die Anmeldung abgelehnt. Bitte Benutzername, Passwort oder Token prüfen.',
  // The server sent an unexpected answer.
  srv_cc10ad5b6bd86945: 'Der Server hat unerwartet geantwortet.',
  // The ICS URL must start with https://, http:// or webcal://.
  srv_457b52382d260269: 'Die ICS-URL muss mit https://, http:// oder webcal:// beginnen.',
  // The URL must start with http:// or https://.
  srv_33e6dad254bc6264: 'Die URL muss mit http:// oder https:// beginnen.',
  // The Wake-on-LAN relay did not answer. Is COMPOSE_PROFILES=wol set (see README)?
  srv_7bd9cc89bdd1560a:
    'Das Wake-on-LAN-Relay hat nicht geantwortet. Ist COMPOSE_PROFILES=wol gesetzt (siehe README)?',
  // The background image must be an http(s) URL.
  srv_28004c9379be9be3: 'Das Hintergrundbild muss eine http(s)-URL sein.',
  // The guest does not exist.
  srv_23575e607ab8d86b: 'Der Gast existiert nicht.',
  // The name must be at most 40 characters long.
  srv_a81a70289d9d44d1: 'Der Name darf höchstens 40 Zeichen lang sein.',
  // The name must be between 1 and 60 characters long.
  srv_71604d59e2455245: 'Der Name muss zwischen 1 und 60 Zeichen lang sein.',
  // Username and app password are required.
  srv_8fe21afcca496dfb: 'Benutzername und App-Passwort sind erforderlich.',
  // Wake-on-LAN is disabled in the demo.
  srv_eceee41d79cdfdb2: 'Wake-on-LAN ist in der Demo deaktiviert.',
  // Weekdays are 0 (Sunday) to 6.
  srv_40623194d868a5f0: 'Wochentage sind 0 (Sonntag) bis 6.',
  // Use a single emoji.
  srv_61031ab185b65e33: 'Bitte nur ein einzelnes Emoji verwenden.',
  // The server answered {status}.
  srv_63f60053f8b51722: 'Der Server antwortete mit {status}.',
};
