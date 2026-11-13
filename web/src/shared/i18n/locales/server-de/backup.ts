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
  // The backup was made by a newer version of trisse-home. Update first, then restore it.
  srv_5b20a69f346d31d3:
    'Das Backup stammt von einer neueren trisse-home-Version. Bitte zuerst aktualisieren und dann wiederherstellen.',
  // The file is not a trisse-home backup.
  srv_2dd869ff96b4bc7a: 'Die Datei ist kein trisse-home-Backup.',
  // The source must be local, remote or upload.
  srv_7054f9e52efbae1e: 'Die Quelle muss local, remote oder upload sein.',
  // User name and password (or app password) are required.
  srv_1c73ff4160f4ee94: 'Benutzername und Passwort (oder App-Passwort) sind erforderlich.',
};
