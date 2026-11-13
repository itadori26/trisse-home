/** Connection settings as shown in the UI (the token secret is never returned). */
export interface ProxmoxSettings {
  configured: boolean;
  url: string;
  tokenId: string;
  hasSecret: boolean;
  fingerprint: string;
}

/** A cluster node; sizes in bytes. */
export interface ProxmoxNode {
  name: string;
  online: boolean;
  cpuPercent: number;
  cpus: number;
  memUsed: number;
  memTotal: number;
  diskUsed: number;
  diskTotal: number;
  uptimeSeconds: number;
}

/** A VM (qemu) or LXC container. */
export interface ProxmoxGuest {
  id: number;
  name: string;
  node: string;
  kind: 'qemu' | 'lxc';
  status: string;
  uptimeSeconds: number;
  cpuPercent: number;
  memUsed: number;
  memTotal: number;
  /** Start and shutdown are allowed (PROXMOX_ACTIONS_GUESTS). */
  actions: boolean;
}

/** State of the cluster. */
export interface ProxmoxStatus {
  nodes: ProxmoxNode[];
  guests: ProxmoxGuest[];
  error: string;
}

/** Power actions; a clean shutdown, never a hard stop. */
export type GuestAction = 'start' | 'shutdown';

/** Share of a total in percent (0 when the total is unknown). */
export function usage(used: number, total: number): number {
  return total > 0 ? (used / total) * 100 : 0;
}

/** The action a guest offers in its current state, if any. */
export function availableAction(g: Pick<ProxmoxGuest, 'actions' | 'status'>): GuestAction | null {
  if (!g.actions) return null;
  if (g.status === 'running') return 'shutdown';
  if (g.status === 'stopped') return 'start';
  return null;
}

/** Widget configuration. */
export interface ProxmoxWidgetConfig {
  /** Only list guests that are running. */
  runningOnly: boolean;
}

/** Reads the widget configuration. */
export function parseWidgetConfig(raw: Record<string, unknown>): ProxmoxWidgetConfig {
  return { runningOnly: raw['runningOnly'] === true };
}

/** Applies the widget configuration. */
export function visibleGuests(list: ProxmoxGuest[], cfg: ProxmoxWidgetConfig): ProxmoxGuest[] {
  return cfg.runningOnly ? list.filter((g) => g.status === 'running') : list;
}
