import type { ServerMessageKey } from '../../../api-client/generated/server-messages/study';

/**
 * German translations of the backend error messages of `study` (#57). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A passing grade lies between 1.0 and 4.0.
  srv_7093d7a65d86c391: 'Eine bestandene Note liegt zwischen 1,0 und 4,0.',
  // At most 200 modules are supported.
  srv_883d152943ffaea6: 'Es werden höchstens 200 Module unterstützt.',
  // Choose open, in progress or passed.
  srv_b8c816562d5fb309: 'Wähle offen, läuft oder bestanden.',
  // ECTS must be between 0.5 and 60 in steps of 0.5.
  srv_8924ee0e3c41840e: 'ECTS müssen zwischen 0,5 und 60 in Schritten von 0,5 liegen.',
  // Enter a module name of at most 80 characters.
  srv_bb48a3c1270ed418: 'Gib einen Modulnamen mit höchstens 80 Zeichen ein.',
  // Only passed modules have a grade.
  srv_488cce48438de374: 'Nur bestandene Module haben eine Note.',
  // The module does not exist.
  srv_9e4c1266ed71da7b: 'Das Modul gibt es nicht.',
  // The semester must be between 1 and 20 (or empty).
  srv_10b16244e1d3f16b: 'Das Semester muss zwischen 1 und 20 liegen (oder leer sein).',
  // The total must be between 1 and 600 ECTS.
  srv_3cf2faced978b8cd: 'Die Gesamtzahl muss zwischen 1 und 600 ECTS liegen.',
};
