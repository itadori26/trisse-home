import { t } from '../../../shared/i18n';
import { normaliseTrackingNumber, type Parcel, type ParcelSettings } from '../domain/parcel';

/** Output port: parcels backend. */
export interface ParcelsGateway {
  settings(): Promise<ParcelSettings>;
  saveSettings(dhlApiKey: string, hideDeliveredDays: number): Promise<ParcelSettings>;
  list(): Promise<Parcel[]>;
  add(trackingNumber: string, carrier: 'dhl', label: string): Promise<Parcel>;
  remove(id: string): Promise<void>;
}

/** Application logic of the parcels feature. */
export class ParcelsUseCases {
  constructor(private readonly gateway: ParcelsGateway) {}

  settings(): Promise<ParcelSettings> {
    return this.gateway.settings();
  }
  saveSettings(dhlApiKey: string, hideDeliveredDays: number): Promise<ParcelSettings> {
    return this.gateway.saveSettings(dhlApiKey.trim(), hideDeliveredDays);
  }
  list(): Promise<Parcel[]> {
    return this.gateway.list();
  }
  /** Validates and adds a DHL parcel. */
  add(trackingNumber: string, label: string): Promise<Parcel> {
    const n = normaliseTrackingNumber(trackingNumber);
    if (!n) return Promise.reject(new Error(t('parcels.invalid_number')));
    return this.gateway.add(n, 'dhl', label.trim());
  }
  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }
}
