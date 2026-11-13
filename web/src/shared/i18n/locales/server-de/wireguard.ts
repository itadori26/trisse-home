import type { ServerMessageKey } from '../../../api-client/generated/server-messages/wireguard';

/**
 * German translations of the backend error messages of `wireguard` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // No WireGuard export found. Run scripts/wireguard-export.sh on the host (see README).
  srv_9d762d7de49b8d86:
    'Kein WireGuard-Export gefunden. Führe scripts/wireguard-export.sh auf dem Host aus (siehe README).',
  // This is not a WireGuard public key.
  srv_e5774de61b49a80f: 'Das ist kein öffentlicher WireGuard-Schlüssel.',
};
