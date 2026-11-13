import type { HttpClient } from '../../../../shared/api-client/http';
import type { BriefingDTO, SettingsDTO } from '../../../../shared/api-client/generated/briefing';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/briefing-requests';
import { type Briefing, type BriefingSettings, isSectionKind, type Severity } from '../../domain/briefing';
import type { BriefingGateway } from '../../usecase/briefingUseCases';

const severity = (s: string): Severity => (s === 'critical' || s === 'warning' ? s : 'info');

const toBriefing = (d: BriefingDTO): Briefing => ({
  sections: d.sections.flatMap((s) =>
    isSectionKind(s.kind)
      ? [
          {
            kind: s.kind,
            more: s.more,
            items: s.items.map((i) => ({
              text: i.text,
              detail: i.detail,
              at: i.at ? new Date(i.at) : null,
              severity: severity(i.severity),
            })),
          },
        ]
      : [],
  ),
  unavailable: (d.unavailable ?? []).filter(isSectionKind),
});

const toSettings = (d: SettingsDTO): BriefingSettings => ({
  layout: d.layout.flatMap((e) => (isSectionKind(e.kind) ? [{ kind: e.kind, enabled: e.enabled }] : [])),
  pushOn: d.pushOn,
  pushAt: d.pushAt,
  pushDays: d.pushDays ?? [],
  skipEmpty: d.skipEmpty,
  lat: d.lat,
  lon: d.lon,
  available: (d.available ?? []).filter(isSectionKind),
});

/** BriefingGateway backed by the REST API. */
export class BriefingApi implements BriefingGateway {
  constructor(private readonly http: HttpClient) {}

  async get(): Promise<Briefing> {
    return toBriefing(await this.http.get<BriefingDTO>('/api/briefing'));
  }

  async settings(): Promise<BriefingSettings> {
    return toSettings(await this.http.get<SettingsDTO>('/api/briefing/settings'));
  }

  async saveSettings(s: BriefingSettings): Promise<BriefingSettings> {
    const body: SettingsInputDTO = {
      layout: s.layout.map((e) => ({ kind: e.kind, enabled: e.enabled })),
      pushOn: s.pushOn,
      pushAt: s.pushAt,
      pushDays: s.pushDays,
      skipEmpty: s.skipEmpty,
      lat: s.lat,
      lon: s.lon,
    };
    return toSettings(await this.http.put<SettingsDTO>('/api/briefing/settings', body));
  }

  sendNow(): Promise<void> {
    return this.http.post('/api/briefing/send');
  }
}
