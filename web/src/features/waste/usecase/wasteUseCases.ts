import {
  validateCalendarFile,
  type WasteOverview,
  type WasteRule,
  type WasteSettings,
} from '../domain/waste';

/** Output port: waste backend. */
export interface WasteGateway {
  settings(): Promise<WasteSettings>;
  saveSettings(url: string, rules: WasteRule[]): Promise<WasteSettings>;
  uploadCalendar(file: File): Promise<WasteSettings>;
  removeCalendar(): Promise<WasteSettings>;
  upcoming(): Promise<WasteOverview>;
}

/** Application logic of the waste feature. */
export class WasteUseCases {
  constructor(private readonly gateway: WasteGateway) {}

  settings(): Promise<WasteSettings> {
    return this.gateway.settings();
  }

  /** Checks and uploads an ICS file as the source; the rules stay (#153). */
  uploadCalendar(file: File): Promise<WasteSettings> {
    const problem = validateCalendarFile(file);
    if (problem) return Promise.reject(new Error(problem));
    return this.gateway.uploadCalendar(file);
  }

  /** Removes the uploaded file; the source is reset, the rules stay. */
  removeCalendar(): Promise<WasteSettings> {
    return this.gateway.removeCalendar();
  }

  /**
   * Saves the rules and, with a URL, switches to that subscription; an empty
   * URL keeps the current source (e.g. an uploaded file). Rules without match
   * text are dropped.
   */
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
