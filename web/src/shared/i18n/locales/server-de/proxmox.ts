import type { ServerMessageKey } from '../../../api-client/generated/server-messages/proxmox';

/**
 * German translations of the backend error messages of `proxmox` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // A certificate fingerprint needs an https:// URL.
  srv_56aacb20f1f37331: 'Ein Zertifikats-Fingerabdruck braucht eine https://-URL.',
  // Actions are not enabled for this guest (PROXMOX_ACTIONS_GUESTS).
  srv_90d6536f1a9bc239: 'Aktionen sind für diesen Gast nicht aktiviert (PROXMOX_ACTIONS_GUESTS).',
  // Proxmox VE is not configured yet.
  srv_d30d28cec1fe8a0e: 'Proxmox VE ist noch nicht eingerichtet.',
  // Proxmox VE is not reachable.
  srv_98db8603c1a1c4cd: 'Proxmox VE ist nicht erreichbar.',
  // The fingerprint must be a SHA-256 value such as AB:CD:…, 32 bytes.
  srv_53b624415dd21efb: 'Der Fingerabdruck muss ein SHA-256-Wert wie AB:CD:… mit 32 Bytes sein.',
  // The token ID must look like monitor@pve!trisse-home.
  srv_98a0961ed24f0281: 'Die Token-ID muss wie monitor@pve!trisse-home aussehen.',
  // The token secret is required.
  srv_3d3a9008e1d23718: 'Das Token-Secret ist erforderlich.',
  // Unknown action, use start or shutdown.
  srv_048731d99c865141: 'Unbekannte Aktion, verwende start oder shutdown.',
};
