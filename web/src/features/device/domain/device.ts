/** How the appliance is used (#140); "" = not chosen yet. */
export type DisplayMode = '' | 'headless' | 'kiosk';

/** Display state of the appliance. */
export interface DeviceState {
  requested: DisplayMode;
  phase: 'idle' | 'pending' | 'installing' | 'active' | 'failed';
  error: string;
  display: 'connected' | 'disconnected' | 'unknown';
}

/** Whether the system is still working on a change (poll until it is not). */
export function changing(s: DeviceState): boolean {
  return s.phase === 'pending' || s.phase === 'installing';
}
