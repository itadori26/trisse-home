import type { WasteOverview, WasteRule, WasteSettings } from '../domain/waste';

/** Output port: waste backend. */
export interface WasteGateway {
  settings(): Promise<WasteSettings>;
  saveSettings(url: string, rules: WasteRule[]): Promise<WasteSettings>;
  upcoming(): Promise<WasteOverview>;
}

/** Application logic of the waste feature. */
export class WasteUseCases {
  constructor(private readonly gateway: WasteGateway) {}

  settings(): Promise<WasteSettings> {
    return this.gateway.settings();
  }

  /** Saves the subscription; rules without match text are dropped. */
  saveSettings(url: string, rules: WasteRule[]): Promise<WasteSettings> {
    const clean = rules
      .map((r) => ({
        match: r.match.trim(),
        label: r.label.trim(),
        color: r.color.trim(),
        icon: r.icon.trim(),
      }))
      .filter((r) => r.match !== '');
    return this.gateway.saveSettings(url.trim(), clean);
  }

  upcoming(): Promise<WasteOverview> {
    return this.gateway.upcoming();
  }
}
