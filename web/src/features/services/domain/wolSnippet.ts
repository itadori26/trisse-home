/** How magic packets reach the LAN. */
export type WolSetup = 'relay' | 'host-network';

/** Input of the .env snippet for Wake-on-LAN. */
export interface WolSnippetInput {
  setup: WolSetup;
  /** Optional directed broadcast such as 192.168.178.255:9. */
  broadcast: string;
  /** Keep the Docker actions profile when it is used as well. */
  dockerActions: boolean;
}

const broadcastPattern = /^(\d{1,3}\.){3}\d{1,3}:\d{1,5}$/;

/** Validates the optional broadcast address (empty is fine). */
export function validBroadcast(broadcast: string): boolean {
  const b = broadcast.trim();
  if (b === '') return true;
  if (!broadcastPattern.test(b)) return false;
  const [ip = '', port = ''] = b.split(':');
  return ip.split('.').every((o) => Number(o) <= 255) && Number(port) > 0 && Number(port) <= 65535;
}

/** Builds the lines to add to .env; null for an invalid broadcast address. */
export function wolEnvSnippet(input: WolSnippetInput): string | null {
  if (!validBroadcast(input.broadcast)) return null;
  const lines: string[] = [];
  if (input.setup === 'relay') {
    lines.push(`COMPOSE_PROFILES=${input.dockerActions ? 'wol,docker-actions' : 'wol'}`);
  } else {
    lines.push('WOL_RELAY=');
  }
  if (input.broadcast.trim() !== '') lines.push(`WOL_BROADCAST=${input.broadcast.trim()}`);
  return lines.join('\n');
}
