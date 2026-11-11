/** Normalised shipment state. */
export type ParcelStatus = 'unknown' | 'pre-transit' | 'transit' | 'delivered' | 'failure';

/** A tracked parcel. */
export interface Parcel {
  id: string;
  trackingNumber: string;
  carrier: 'dhl';
  label: string;
  status: ParcelStatus;
  description: string;
  location: string;
  statusAt: Date | null;
  estimatedDelivery: Date | null;
  error: string;
}

/** Settings as shown in the UI (the key is never returned). */
export interface ParcelSettings {
  hasDhlApiKey: boolean;
  hideDeliveredDays: number;
}

/** Short label and color per status. */
export const statusInfo: Record<ParcelStatus, { label: string; color: string }> = {
  unknown: { label: 'Unknown', color: 'var(--color-text-muted)' },
  'pre-transit': { label: 'Announced', color: 'var(--color-text-muted)' },
  transit: { label: 'On the way', color: 'var(--color-accent)' },
  delivered: { label: 'Delivered', color: 'var(--color-success)' },
  failure: { label: 'Problem', color: 'var(--color-danger)' },
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
