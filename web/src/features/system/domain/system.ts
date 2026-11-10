/** Used/total pair in bytes. */
export interface Usage {
  total: number;
  used: number;
  percent: number;
}

/** Host metrics snapshot. */
export interface SystemSnapshot {
  cpuPercent: number;
  cpuCount: number;
  load: [number, number, number];
  memory: Usage;
  swap: Usage;
  disks: { label: string; usage: Usage }[];
  uptimeSeconds: number;
  temperature: number;
}

/** Normalised load: 1-minute load per CPU in percent (100 = fully busy). */
export function loadPercent(s: SystemSnapshot): number {
  if (s.cpuCount <= 0) return 0;
  return Math.min(100, (s.load[0] / s.cpuCount) * 100);
}
