import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  MailSettingsDTO,
  ParcelDTO,
  SettingsDTO,
} from '../../../../shared/api-client/generated/parcels';
import type {
  MailSettingsInputDTO,
  NewParcelDTO,
  SettingsInputDTO,
} from '../../../../shared/api-client/generated/parcels-requests';
import type { Carrier, MailSettings, MailSettingsInput, Parcel, ParcelSettings } from '../../domain/parcel';
import type { ParcelsGateway } from '../../usecase/parcelsUseCases';

const toParcel = (d: ParcelDTO): Parcel => ({
  ...d,
  statusAt: d.statusAt ? new Date(d.statusAt) : null,
  estimatedDelivery: d.estimatedDelivery ? new Date(d.estimatedDelivery) : null,
});

const toMail = (d: MailSettingsDTO): MailSettings => ({
  ...d,
  lastRun: d.lastRun ? new Date(d.lastRun) : null,
});

/** ParcelsGateway backed by the REST API. */
export class ParcelsApi implements ParcelsGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<ParcelSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/parcels/settings')) };
  }
  async saveSettings(dhlApiKey: string, hideDeliveredDays: number): Promise<ParcelSettings> {
    const body: SettingsInputDTO = { dhlApiKey, hideDeliveredDays };
    return { ...(await this.http.put<SettingsDTO>('/api/parcels/settings', body)) };
  }
  async list(): Promise<Parcel[]> {
    return (await this.http.get<ParcelDTO[]>('/api/parcels')).map(toParcel);
  }
  async add(trackingNumber: string, carrier: Carrier, label: string): Promise<Parcel> {
    const body: NewParcelDTO = { trackingNumber, carrier, label };
    return toParcel(await this.http.post<ParcelDTO>('/api/parcels', body));
  }
  remove(id: string): Promise<void> {
    return this.http.delete(`/api/parcels/${encodeURIComponent(id)}`);
  }
  async mailSettings(): Promise<MailSettings> {
    return toMail(await this.http.get<MailSettingsDTO>('/api/parcels/mail'));
  }
  async saveMailSettings(input: MailSettingsInput): Promise<MailSettings> {
    const { enabled, host, port, username, password, folder, intervalMinutes } = input;
    const body: MailSettingsInputDTO = { enabled, host, port, username, password, folder, intervalMinutes };
    return toMail(await this.http.put<MailSettingsDTO>('/api/parcels/mail', body));
  }
}
