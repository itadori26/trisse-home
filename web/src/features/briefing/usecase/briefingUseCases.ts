import { type Briefing, type BriefingSettings, validateSettings } from '../domain/briefing';

/** Output port: briefing backend. */
export interface BriefingGateway {
  get(): Promise<Briefing>;
  settings(): Promise<BriefingSettings>;
  saveSettings(s: BriefingSettings): Promise<BriefingSettings>;
  sendNow(): Promise<void>;
}

/** Error for input rejected before it reaches the backend. */
export class BriefingValidationError extends Error {
  override name = 'BriefingValidationError';
}

/** Application logic of the briefing feature (#52, #70). */
export class BriefingUseCases {
  constructor(private readonly gateway: BriefingGateway) {}

  get(): Promise<Briefing> {
    return this.gateway.get();
  }

  settings(): Promise<BriefingSettings> {
    return this.gateway.settings();
  }

  /** Saves layout, push schedule and location after validation. */
  async saveSettings(s: BriefingSettings): Promise<BriefingSettings> {
    const problem = validateSettings(s);
    if (problem) throw new BriefingValidationError(problem);
    return this.gateway.saveSettings({ ...s, pushDays: [...new Set(s.pushDays)].sort() });
  }

  /** Pushes the briefing right away (test). */
  sendNow(): Promise<void> {
    return this.gateway.sendNow();
  }
}
