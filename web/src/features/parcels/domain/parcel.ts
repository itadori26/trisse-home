import type { MessageKey } from '../../../shared/i18n';

/** Normalised shipment state. */
export type ParcelStatus = 'unknown' | 'pre-transit' | 'transit' | 'delivered' | 'failure';

/** Carriers; only DHL has live tracking, the others link to their page (#166). */
export type Carrier = 'dhl' | 'hermes' | 'dpd' | 'gls' | 'ups';
export const CARRIERS: readonly Carrier[] = ['dhl', 'hermes', 'dpd', 'gls', 'ups'];

/** Display name of a carrier. */
export function carrierName(c: Carrier): string {
  return { dhl: 'DHL', hermes: 'Hermes', dpd: 'DPD', gls: 'GLS', ups: 'UPS' }[c];
}

/** A tracked parcel. */
export interface Parcel {
  id: string;
  trackingNumber: string;
  carrier: Carrier;
  label: string;
  status: ParcelStatus;
  description: string;
  location: string;
  statusAt: Date | null;
  estimatedDelivery: Date | null;
  error: string;
  /** Found in a shipping e-mail (#166). */
  discovered: boolean;
  /** The carrier's public tracking page. */
  trackingUrl: string;
  /** False for carriers without live tracking: only the link. */
  trackable: boolean;
}

/** The read-only mailbox parcels are found in (#166); never the password. */
export interface MailSettings {
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  hasPassword: boolean;
  folder: string;
  intervalMinutes: number;
  lastRun: Date | null;
  lastError: string;
  found: number;
}

/** Input of saving the mailbox; an empty password keeps the stored one. */
export interface MailSettingsInput {
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  password: string;
  folder: string;
  intervalMinutes: number;
}

/** Settings as shown in the UI (the key is never returned). */
export interface ParcelSettings {
  hasDhlApiKey: boolean;
  hideDeliveredDays: number;
}

/** Short label and color per status. */
export const statusInfo: Record<ParcelStatus, { label: MessageKey; color: string }> = {
  unknown: { label: 'parcels.status.unknown', color: 'var(--color-text-muted)' },
  'pre-transit': { label: 'parcels.status.pre_transit', color: 'var(--color-text-muted)' },
  transit: { label: 'parcels.status.transit', color: 'var(--color-accent)' },
  delivered: { label: 'parcels.status.delivered', color: 'var(--color-success)' },
  failure: { label: 'parcels.status.failure', color: 'var(--color-danger)' },
};

/** Removes spaces and validates a tracking number (8–35 letters/digits). */
export function normaliseTrackingNumber(s: string): string | null {
  const n = s.replace(/\s+/g, '').toUpperCase();
  return /^[A-Z0-9]{8,35}$/.test(n) ? n : null;
}

/** Display name: label or the tracking number. */
export function parcelName(p: Pick<Parcel, 'label' | 'trackingNumber'>): string {
  return p.label || p.trackingNumber;
}
