import type { ServerMessageKey } from '../../../api-client/generated/server-messages/dashboard';

/**
 * German translations of the backend error messages of `dashboard` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most 10 rules per tab are supported.
  srv_9c62e19aa7e4b027: 'Pro Tab werden höchstens 10 Regeln unterstützt.',
  // At most 50 templates can be stored.
  srv_fb588d9d1c9ff8e5: 'Es können höchstens 50 Vorlagen gespeichert werden.',
  // Every widget may appear only once.
  srv_cd8384030060ee99: 'Jedes Widget darf nur einmal vorkommen.',
  // Rules need weekdays 0–6, times like 07:30 and networks like 192.168.178.0/24.
  srv_560f31f8940f8cfc: 'Regeln brauchen Wochentage 0–6, Uhrzeiten wie 07:30 und Netze wie 192.168.178.0/24.',
  // The height must be between 1 and 24 rows.
  srv_0d5b9b677899722e: 'Die Höhe muss zwischen 1 und 24 Zeilen liegen.',
  // The last tab cannot be deleted.
  srv_fd5437e3a5a60614: 'Der letzte Tab kann nicht gelöscht werden.',
  // The new order must contain every tab exactly once.
  srv_3364239a85507862: 'Die neue Reihenfolge muss jeden Tab genau einmal enthalten.',
  // The row is out of range.
  srv_088ff184c130f97e: 'Die Zeile liegt außerhalb des Bereichs.',
  // The tab does not exist.
  srv_7e608cff024ed2a8: 'Der Tab existiert nicht.',
  // The tab has too many widgets for a template.
  srv_12af04b41c068c5d: 'Der Tab hat zu viele Widgets für eine Vorlage.',
  // The tab name must be at most 40 characters long.
  srv_e24dab6aca94e3e8: 'Der Tab-Name darf höchstens 40 Zeichen lang sein.',
  // The tab name must not be empty.
  srv_31748b4a01624458: 'Der Tab-Name darf nicht leer sein.',
  // The template does not exist.
  srv_4ad756b3998ef1fd: 'Die Vorlage existiert nicht.',
  // The template name must be between 1 and 40 characters long.
  srv_c62fe8c4040fbb11: 'Der Vorlagenname muss zwischen 1 und 40 Zeichen lang sein.',
  // The widget configuration is too large.
  srv_3f4f8e97996c49f7: 'Die Widget-Konfiguration ist zu groß.',
  // The widget configuration must be a JSON object.
  srv_90febdf2f4113f20: 'Die Widget-Konfiguration muss ein JSON-Objekt sein.',
  // The widget does not exist.
  srv_cff3a20eb39a88df: 'Das Widget existiert nicht.',
  // The widget must fit into the 12-column grid.
  srv_fbdc8d1ab5c933c9: 'Das Widget muss in das 12-spaltige Raster passen.',
  // The widget type is invalid.
  srv_a164605f887b18d7: 'Der Widget-Typ ist ungültig.',
  // The width must be between 1 and 12 columns.
  srv_9dfec97d8d360b86: 'Die Breite muss zwischen 1 und 12 Spalten liegen.',
  // This widget type is not available.
  srv_63543557e19be101: 'Dieser Widget-Typ ist nicht verfügbar.',
};
