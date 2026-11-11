/** Input of the cron snippet for the WireGuard export script. */
export interface ExportSnippetInput {
  /** Directory of the trisse-home checkout on the host. */
  stackDir: string;
  /** Container running WireGuard; empty when wg runs on the host. */
  container: string;
}

const shellSafe = /^[A-Za-z0-9_./-]*$/;

/**
 * Builds the /etc/cron.d entry that exports the peer status every minute.
 * Returns null for input that is not a plain path/name (no shell injection
 * through a copied snippet).
 */
export function exportCronSnippet(input: ExportSnippetInput): string | null {
  const dir = input.stackDir.trim().replace(/\/+$/, '');
  const container = input.container.trim();
  if (!dir.startsWith('/') || !shellSafe.test(dir) || !shellSafe.test(container)) return null;
  const env = container === 'wireguard' ? '' : `WG_CONTAINER=${container} `;
  return [
    `sudo tee /etc/cron.d/trisse-home-wireguard <<'CRON'`,
    `* * * * * root ${env}${dir}/scripts/wireguard-export.sh`,
    'CRON',
  ].join('\n');
}
