/** Health state of a service. */
export type ServiceState = 'online' | 'offline' | 'unknown';

/** Editable fields of a service tile. */
export interface ServiceInput {
  name: string;
  url: string;
  checkUrl: string;
  icon: string;
  description: string;
  skipTlsVerify: boolean;
}

/** A service tile with its latest health check. */
export interface Service extends ServiceInput {
  id: string;
  position: number;
  state: ServiceState;
  httpStatus: number;
  latencyMs: number;
  checkedAt: Date | null;
  error: string;
  certExpiresAt: Date | null;
  certExpiresSoon: boolean;
}

/** One point of a service's response time history. */
export interface LatencyPoint {
  at: Date;
  latencyMs: number;
  online: boolean;
}

/** Whole days until the TLS certificate expires, or null without certificate. */
export function certificateDaysLeft(s: Pick<Service, 'certExpiresAt'>, now: Date): number | null {
  if (!s.certExpiresAt) return null;
  return Math.floor((s.certExpiresAt.getTime() - now.getTime()) / 86_400_000);
}

/** An empty form. */
export const emptyServiceInput = (): ServiceInput => ({
  name: '',
  url: '',
  checkUrl: '',
  icon: '',
  description: '',
  skipTlsVerify: false,
});

/** Reports whether the icon should be rendered as an image. */
export function isImageIcon(icon: string): boolean {
  return /^https?:\/\//i.test(icon);
}

/** Short status text for tooltips. */
export function statusText(s: Service): string {
  if (s.state === 'unknown') return 'Not checked yet';
  if (s.state === 'offline') return s.error ? `Offline: ${s.error}` : `Offline (HTTP ${s.httpStatus})`;
  return `Online · HTTP ${s.httpStatus} · ${s.latencyMs} ms`;
}

/** Validates the form before it is sent and returns an error message or null. */
export function validateServiceInput(input: ServiceInput): string | null {
  if (input.name.trim() === '') return 'Please enter a name.';
  if (!/^https?:\/\/\S+$/i.test(input.url.trim())) return 'The URL must start with http:// or https://.';
  if (input.checkUrl.trim() !== '' && !/^https?:\/\/\S+$/i.test(input.checkUrl.trim()))
    return 'The check URL must start with http:// or https://.';
  return null;
}
