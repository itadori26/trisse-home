import { locale } from '../../../shared/i18n';
import type { Summarizer, SummaryOutput } from '../../../shared/ui/summarizer';
import type { SummarySettings } from '../domain/summary';

/** Output port: summary backend. */
export interface SummaryGateway {
  settings(): Promise<SummarySettings>;
  saveSettings(url: string, model: string, maxInputChars: number): Promise<SummarySettings>;
  disable(): Promise<void>;
  models(url: string): Promise<string[]>;
  summarize(input: { text?: string; url?: string; language: string }): Promise<SummaryOutput>;
}

/** Application logic of the summaries (#75); also the Summarizer of notes and feeds. */
export class SummaryUseCases implements Summarizer {
  private status: Promise<boolean> | null = null;

  constructor(private readonly gateway: SummaryGateway) {}

  /** Whether Ollama is set up; asked once, refreshed after saving. */
  available(): Promise<boolean> {
    this.status ??= this.gateway.settings().then(
      (s) => s.configured,
      () => false,
    );
    return this.status;
  }
  text(text: string): Promise<SummaryOutput> {
    return this.gateway.summarize({ text, language: locale() });
  }
  link(url: string): Promise<SummaryOutput> {
    return this.gateway.summarize({ url, language: locale() });
  }
  settings(): Promise<SummarySettings> {
    return this.gateway.settings();
  }
  async saveSettings(url: string, model: string, maxInputChars: number): Promise<SummarySettings> {
    const s = await this.gateway.saveSettings(url.trim(), model, maxInputChars);
    this.status = Promise.resolve(s.configured);
    return s;
  }
  async disable(): Promise<void> {
    await this.gateway.disable();
    this.status = Promise.resolve(false);
  }
  models(url: string): Promise<string[]> {
    return this.gateway.models(url.trim());
  }
}
