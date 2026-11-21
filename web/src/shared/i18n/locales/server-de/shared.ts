import type { ServerMessageKey } from '../../../api-client/generated/server-messages/shared';

/**
 * German translations of the backend error messages of `shared` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Enter a date like 2026-12-24.
  srv_22fc413ef29251fa: 'Bitte ein Datum wie 2026-12-24 eingeben.',
  // Nothing answers at this address. Is the service running?
  srv_8a17d4cf0faa9d18: 'Unter dieser Adresse antwortet nichts. Läuft der Dienst?',
  // The certificate is not trusted. Use the http:// address in your home network, or give the service a certificate this server trusts.
  srv_32f1f2343e831224:
    'Dem Zertifikat wird nicht vertraut. Nutze im Heimnetz die http://-Adresse oder gib dem Dienst ein Zertifikat, dem dieser Server vertraut.',
  // The key is not accepted. Was it copied completely?
  srv_48527dc59bd9b8f6: 'Der Schlüssel wird nicht akzeptiert. Wurde er vollständig kopiert?',
  // This address does not speak https. Try it with http://.
  srv_b22ef860827fa826: 'Diese Adresse spricht kein https. Versuche es mit http://.',
  // This is not {service}. Something else runs at this address.
  srv_3cfc83de58917c87: 'Das ist nicht {service}. Unter dieser Adresse läuft etwas anderes.',
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
  // The Wake-on-LAN relay did not answer. Is COMPOSE_PROFILES=wol set (see docs/host-integrations.md)?
  srv_39f1fbfcd5e03469:
    'Das Wake-on-LAN-Relay hat nicht geantwortet. Ist COMPOSE_PROFILES=wol gesetzt (siehe docs/host-integrations.md)?',
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
  // The ID must have 8 to 64 letters, digits, '-' or '_'.
  srv_9cdf89c6055b9583: "Die ID muss aus 8 bis 64 Buchstaben, Ziffern, '-' oder '_' bestehen.",
  // User name and app password are required.
  srv_970344b28687ee67: 'Benutzername und App-Passwort sind erforderlich.',
  // Enter the IMAP server name, e.g. imap.example.com.
  srv_91809f90e8dfc4fb: 'Gib den Namen des IMAP-Servers ein, z. B. imap.example.com.',
  // The port must be between 1 and 65535 (IMAP over TLS: 993).
  srv_92285ba58a255b47: 'Der Port muss zwischen 1 und 65535 liegen (IMAP über TLS: 993).',
  // User name, password and folder must not contain line breaks.
  srv_b08dda17b8ef851b: 'Benutzername, Passwort und Ordner dürfen keine Zeilenumbrüche enthalten.',
  // These settings are managed by the configuration file (CONFIG_FILE) and are read-only here.
  srv_07a824b208b96b62:
    'Diese Einstellungen werden von der Konfigurationsdatei (CONFIG_FILE) verwaltet und sind hier schreibgeschützt.',
  // The background must be an http(s) image URL, an uploaded picture or a built-in background.
  srv_fc9cd325f2b0b2ac:
    'Der Hintergrund muss eine http(s)-Bildadresse, ein hochgeladenes Bild oder ein eingebauter Hintergrund sein.',
  // Upload a JPEG, PNG or GIF picture.
  srv_eac41835716f01a3: 'Lade ein JPEG-, PNG- oder GIF-Bild hoch.',
  // Enter the client ID of your app registration.
  srv_c8433603152a8aec: 'Gib die Client-ID deiner App-Registrierung ein.',
  // Signing in needs PUBLIC_URL (an https:// address of Waylight).
  srv_44cddb4885b88020: 'Zum Anmelden muss PUBLIC_URL gesetzt sein (eine https://-Adresse von Waylight).',
  // The sign-in has expired or was not finished. Please sign in again.
  srv_88824ac4a9a1e957: 'Die Anmeldung ist abgelaufen oder wurde nicht abgeschlossen. Bitte erneut anmelden.',
  // The sign-in was not started here or has expired. Please start again.
  srv_cd970adb7028a7ce: 'Die Anmeldung wurde nicht hier gestartet oder ist abgelaufen. Bitte neu beginnen.',
  // Too many sign-ins at once. Please wait a few minutes.
  srv_851e9d41e5c4c49f: 'Zu viele Anmeldungen gleichzeitig. Bitte warte ein paar Minuten.',
  // The name must have between 1 and 60 characters.
  srv_9f138e811fd44724: 'Der Name muss zwischen 1 und 60 Zeichen haben.',
  // Only the person who created this can change it.
  srv_6699f3250edaed8e: 'Das kann nur ändern, wer es angelegt hat.',
  // Only an admin can change this.
  srv_e5be262a5868a63b: 'Das kann nur ein Admin ändern.',
  // Share with only you, everyone or one of your groups.
  srv_cacad70ec42d516f: 'Teile nur mit dir, mit allen oder mit einer deiner Gruppen.',
  // Please enter a valid e-mail address.
  srv_6ddc9c2076b9e645: 'Bitte gib eine gültige E-Mail-Adresse ein.',
  // This feature needs a valid licence key.
  srv_b093b38e42d34fa3: 'Diese Funktion braucht einen gültigen Lizenzschlüssel.',
  // The licence key has expired or covers fewer accounts than this instance has; …
  srv_28cc0a1c84155db5:
    'Der Lizenzschlüssel ist abgelaufen oder deckt weniger Konten ab, als diese Instanz hat. Die Daten bleiben lesbar, Änderungen brauchen einen gültigen Schlüssel.',
  // The service does not answer.
  srv_d4a77dfc8f36f361: 'Der Dienst antwortet nicht.',
};
