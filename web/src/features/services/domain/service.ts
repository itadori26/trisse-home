import { t } from '../../../shared/i18n';
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
  /** MAC address for Wake-on-LAN, empty if none. */
  mac: string;
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
  /** Source container of a discovered tile, empty for manual tiles (#78). */
  container: string;
  /** The container is gone: the tile is kept but not shown on the dashboard. */
  hidden: boolean;
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
  mac: '',
});

const macPattern = /^([0-9a-f]{2}[:-]?){5}[0-9a-f]{2}$|^[0-9a-f]{4}\.[0-9a-f]{4}\.[0-9a-f]{4}$/i;

/** Whether the tile offers a Wake-on-LAN button (only while not online). */
export function canWake(s: Pick<Service, 'mac' | 'state'>): boolean {
  return s.mac !== '' && s.state !== 'online';
}

/** Reports whether the icon should be rendered as an image. */
export function isImageIcon(icon: string): boolean {
  return /^https?:\/\//i.test(icon);
}

/** Short status text for tooltips. */
export function statusText(s: Service): string {
  if (s.state === 'unknown') return t('services.not_checked');
  if (s.state === 'offline')
    return s.error
      ? t('services.offline_error', { error: s.error })
      : t('services.offline_http', { status: s.httpStatus });
  return t('services.online', { status: s.httpStatus, ms: s.latencyMs });
}

/** Validates the form before it is sent and returns an error message or null. */
export function validateServiceInput(input: ServiceInput): string | null {
  if (input.name.trim() === '') return t('services.enter_name');
  if (!/^https?:\/\/\S+$/i.test(input.url.trim())) return t('services.invalid_url');
  if (input.checkUrl.trim() !== '' && !/^https?:\/\/\S+$/i.test(input.checkUrl.trim()))
    return t('services.invalid_check_url');
  if (input.mac.trim() !== '' && !macPattern.test(input.mac.trim())) return t('services.invalid_mac');
  return null;
}

/** Tiles shown on the dashboard: discovered tiles of removed containers are hidden. */
export function visibleServices<T extends Pick<Service, 'hidden'>>(list: readonly T[]): T[] {
  return list.filter((s) => !s.hidden);
}
