import type { HttpClient } from '../../../../shared/api-client/http';
import type { ModelsDTO, ResultDTO, SettingsDTO } from '../../../../shared/api-client/generated/summary';
import type {
  AskDTO,
  ModelsInputDTO,
  SettingsInputDTO,
  SummarizeDTO,
} from '../../../../shared/api-client/generated/summary-requests';
import type { SummaryOutput } from '../../../../shared/ui/summarizer';
import type { SummarySettings } from '../../domain/summary';
import type { SummaryGateway } from '../../usecase/summaryUseCases';

/** SummaryGateway backed by the REST API. */
export class SummaryApi implements SummaryGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<SummarySettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/summary/settings')) };
  }
  async saveSettings(url: string, model: string, maxInputChars: number): Promise<SummarySettings> {
    const body: SettingsInputDTO = { url, model, maxInputChars };
    return { ...(await this.http.put<SettingsDTO>('/api/summary/settings', body)) };
  }
  disable(): Promise<void> {
    return this.http.delete('/api/summary/settings');
  }
  async models(url: string): Promise<string[]> {
    const body: ModelsInputDTO = { url };
    return (await this.http.post<ModelsDTO>('/api/summary/models', body)).models ?? [];
  }
  async summarize(input: { text?: string; url?: string; language: string }): Promise<SummaryOutput> {
    const body: SummarizeDTO = input;
    const r = await this.http.post<ResultDTO>('/api/summary', body);
    return { summary: r.summary, truncated: r.truncated, cached: r.cached };
  }
  ask(
    input: { question: string; context: string; language: string },
    signal: AbortSignal,
  ): Promise<AsyncIterable<string>> {
    if (!this.http.postStream) return Promise.reject(new Error('streaming is not supported'));
    const body: AskDTO = input;
    return this.http.postStream('/api/summary/ask', body, signal);
  }
}
