/** Health status of a disk, ordered by severity. */
export type DiskStatus = 'ok' | 'unknown' | 'warning' | 'critical';

/** A rule that affected the status; code + value/name are translated by the UI. */
export interface Finding {
  status: DiskStatus;
  code: string;
  value: number;
  name: string;
  message: string;
}

/** One disk as evaluated by the server (#59). */
export interface Disk {
  device: string;
  model: string;
  protocol: 'ATA' | 'NVMe' | 'other';
  capacityBytes: number;
  checkedAt: Date;
  status: DiskStatus;
  findings: Finding[];
  temperature: number | null;
  powerOnHours: number | null;
  reallocated: number | null;
  pending: number | null;
  percentageUsed: number | null;
}

const order: Record<DiskStatus, number> = { ok: 0, unknown: 1, warning: 2, critical: 3 };

/** The worst status of all disks (ok for none). */
export function overallStatus(disks: readonly Disk[]): DiskStatus {
  return disks.reduce<DiskStatus>((w, d) => (order[d.status] > order[w] ? d.status : w), 'ok');
}

/** Power-on time as years and days, e.g. { years: 3, days: 310 }. */
export function powerOnAge(hours: number): { years: number; days: number } {
  const days = Math.floor(hours / 24);
  return { years: Math.floor(days / 365), days: days % 365 };
}

/** Finding codes known to the UI catalogue; others fall back to the message. */
export const findingCodes = [
  'self_assessment_failed',
  'temperature',
  'attribute_failing',
  'reallocated',
  'reported_uncorrectable',
  'pending',
  'offline_uncorrectable',
  'nvme_critical_warning',
  'spare_low',
  'wear',
  'media_errors',
  'no_data',
  'stale',
] as const;

/** A finding code with a UI translation. */
export type FindingCode = (typeof findingCodes)[number];

/** Reports whether the UI can translate a finding code. */
export function isKnownFinding(code: string): code is FindingCode {
  return (findingCodes as readonly string[]).includes(code);
}
