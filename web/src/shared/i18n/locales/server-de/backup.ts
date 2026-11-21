import type { ServerMessageKey } from '../../../api-client/generated/server-messages/backup';

/**
 * German translations of the backend error messages of `backup` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Between 1 and 365 backups can be kept.
  srv_4bba6d234928bd07: 'Es können zwischen 1 und 365 Backups aufbewahrt werden.',
  // Choose a backup file.
  srv_56aef011f47cdbeb: 'Bitte eine Backup-Datei auswählen.',
  // No off-site backup target is configured.
  srv_7f26dde75768b4ec: 'Es ist kein externes Backup-Ziel eingerichtet.',
  // Only the owner can restore backups in the browser.
  srv_2fa2f3b7606fee5e: 'Nur der Besitzer kann Backups im Browser wiederherstellen.',
  // The backup could not be received: {detail}
  srv_ad9909a46d053d7b: 'Das Backup konnte nicht empfangen werden: {detail}',
  // The backup does not exist.
  srv_1a3955e489757321: 'Das Backup existiert nicht.',
  // The backup interval must be at least one hour.
  srv_0fd566747f1cd503: 'Das Backup-Intervall muss mindestens eine Stunde betragen.',
  // The backup is damaged (integrity check failed).
  srv_678e4dc426de200a: 'Das Backup ist beschädigt (Integritätsprüfung fehlgeschlagen).',
  // The backup was made by a newer version of Waylight. Update first, then restore it.
  srv_c8745160757d198c:
    'Das Backup stammt von einer neueren Waylight-Version. Bitte zuerst aktualisieren und dann wiederherstellen.',
  // The file is not a Waylight backup.
  srv_cc68528bcc0907b5: 'Die Datei ist kein Waylight-Backup.',
  // The source must be local, remote or upload.
  srv_7054f9e52efbae1e: 'Die Quelle muss local, remote oder upload sein.',
  // User name and password (or app password) are required.
  srv_1c73ff4160f4ee94: 'Benutzername und Passwort (oder App-Passwort) sind erforderlich.',
  // Backup older than expected
  srv_6a02ee04f1083e2c: 'Backup älter als erwartet',
  // Backups all good
  srv_d0f41247027c6d48: 'Backups in Ordnung',
  // Failed {time}
  srv_b2df19aea32832fb: 'Fehlgeschlagen {time}',
  // Last backup {time}
  srv_771c68846be5456c: 'Letztes Backup {time}',
  // No backup yet
  srv_9a9b8a508d5dcf82: 'Noch kein Backup',
  // The last backup failed
  srv_0b6056a8bbfbd827: 'Das letzte Backup ist fehlgeschlagen',
  // The off-site copy failed
  srv_4ad084d71a9bd861: 'Die externe Kopie ist fehlgeschlagen',
};
