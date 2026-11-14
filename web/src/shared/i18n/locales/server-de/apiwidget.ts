import type { ServerMessageKey } from '../../../api-client/generated/server-messages/apiwidget';

/**
 * German translations of the backend error messages of `apiwidget` (#48). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Add between 1 and 10 values.
  srv_afd0ddf31c97d37b: 'Lege zwischen 1 und 10 Werte an.',
  // At most 50 endpoints are supported.
  srv_9a6d26e600bc942c: 'Es werden höchstens 50 Endpunkte unterstützt.',
  // Enter a name of at most 60 characters.
  srv_a9c3787f1a8c5e72: 'Gib einen Namen mit höchstens 60 Zeichen ein.',
  // Enter the header value.
  srv_9388f7d78d1421bc: 'Gib den Wert des Headers ein.',
  // Every value needs a JSONPath of at most 200 characters.
  srv_3aa18ca8aec1b0b1: 'Jeder Wert braucht einen JSONPath mit höchstens 200 Zeichen.',
  // Every value needs a label of at most 40 characters.
  srv_ed146afca70b6b18: 'Jeder Wert braucht eine Bezeichnung mit höchstens 40 Zeichen.',
  // Invalid JSONPath {path}: {detail}
  srv_18b0ba190b19d73b: 'Ungültiger JSONPath {path}: {detail}',
  // The address is not in a private network. Allow public addresses for this endpoint to reach it.
  srv_96e158ab8fbd211f:
    'Die Adresse liegt nicht in einem privaten Netz. Erlaube öffentliche Adressen für diesen Endpunkt, um sie zu erreichen.',
  // The endpoint answered {status}.
  srv_719bb42774762c3c: 'Der Endpunkt hat mit {status} geantwortet.',
  // The endpoint did not answer within 10 seconds.
  srv_6e83c7041d0c7d02: 'Der Endpunkt hat nicht innerhalb von 10 Sekunden geantwortet.',
  // The endpoint does not exist.
  srv_e933bf0bb56c202a: 'Den Endpunkt gibt es nicht.',
  // The header name may only contain letters, digits and - _ . characters.
  srv_9866c93cb1f70331: 'Der Header-Name darf nur Buchstaben, Ziffern und - _ . enthalten.',
  // The header value is too long or contains line breaks.
  srv_c337f450f9a6424e: 'Der Header-Wert ist zu lang oder enthält Zeilenumbrüche.',
  // The interval must be between 30 seconds and 24 hours.
  srv_1027da36188a6ee7: 'Das Intervall muss zwischen 30 Sekunden und 24 Stunden liegen.',
  // The response is larger than 1 MB.
  srv_ff7c627f1b5a4717: 'Die Antwort ist größer als 1 MB.',
  // The response is not JSON.
  srv_986fb912129aa8e8: 'Die Antwort ist kein JSON.',
  // This header is set automatically and cannot be configured.
  srv_04481dd132908420: 'Dieser Header wird automatisch gesetzt und kann nicht eingestellt werden.',
  // Thresholds must be numbers.
  srv_6a2a422643619241: 'Schwellwerte müssen Zahlen sein.',
  // Units must be at most 12 characters long.
  srv_fe9ef3e59d3ce637: 'Einheiten dürfen höchstens 12 Zeichen lang sein.',
};
