import { t } from '../../../shared/i18n';
import {
  type Carrier,
  type MailSettings,
  type MailSettingsInput,
  normaliseTrackingNumber,
  type Parcel,
  type ParcelSettings,
} from '../domain/parcel';

/** Output port: parcels backend. */
export interface ParcelsGateway {
  settings(): Promise<ParcelSettings>;
  saveSettings(dhlApiKey: string, hideDeliveredDays: number): Promise<ParcelSettings>;
  list(): Promise<Parcel[]>;
  add(trackingNumber: string, carrier: Carrier, label: string): Promise<Parcel>;
  remove(id: string): Promise<void>;
  mailSettings(): Promise<MailSettings>;
  saveMailSettings(input: MailSettingsInput): Promise<MailSettings>;
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
  /** Validates and adds a parcel (DHL unless another carrier is chosen). */
  add(trackingNumber: string, label: string, carrier: Carrier = 'dhl'): Promise<Parcel> {
    const n = normaliseTrackingNumber(trackingNumber);
    if (!n) return Promise.reject(new Error(t('parcels.invalid_number')));
    return this.gateway.add(n, carrier, label.trim());
  }
  /** The mail discovery settings (#166). */
  mailSettings(): Promise<MailSettings> {
    return this.gateway.mailSettings();
  }
  /** Saves the mailbox; the server tests an enabled one first. */
  saveMailSettings(input: MailSettingsInput): Promise<MailSettings> {
    return this.gateway.saveMailSettings({
      ...input,
      host: input.host.trim(),
      username: input.username.trim(),
      folder: input.folder.trim(),
    });
  }
  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }
}
