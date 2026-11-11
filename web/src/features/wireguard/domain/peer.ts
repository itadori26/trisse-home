/** A WireGuard peer. */
export interface Peer {
  publicKey: string;
  name: string;
  interface: string;
  endpoint: string;
  allowedIps: string;
  online: boolean;
  latestHandshake: Date | null;
  receivedBytes: number;
  sentBytes: number;
}

/** Peers with export metadata. */
export interface PeerList {
  peers: Peer[];
  online: number;
  updatedAt: Date;
  stale: boolean;
}

/** "just now", "5 min ago", "3 h ago", "2 d ago" or "never". */
export function handshakeAgo(d: Date | null, now: Date): string {
  if (!d) return 'never';
  const s = Math.max(0, (now.getTime() - d.getTime()) / 1000);
  if (s < 60) return 'just now';
  if (s < 3600) return `${Math.floor(s / 60)} min ago`;
  if (s < 172_800) return `${Math.floor(s / 3600)} h ago`;
  return `${Math.floor(s / 86_400)} d ago`;
}
