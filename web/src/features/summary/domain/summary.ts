/** Ollama setup as shown in the UI (#75). */
export interface SummarySettings {
  configured: boolean;
  url: string;
  model: string;
  maxInputChars: number;
}
