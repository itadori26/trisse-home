/** A summary of a note or feed article (#75). */
export interface SummaryOutput {
  summary: string;
  /** Only the beginning of the text was summarised. */
  truncated: boolean;
  cached: boolean;
}

/**
 * What notes and the feed reader need from the summary feature. Defined here
 * so features stay independent; src/app passes the summary use cases in.
 */
export interface Summarizer {
  /** False while no Ollama is set up (or the module is disabled). */
  available(): Promise<boolean>;
  text(text: string): Promise<SummaryOutput>;
  link(url: string): Promise<SummaryOutput>;
  /**
   * Answers a free question of the search (#162) from a compact status
   * context; the answer arrives in chunks, aborting stops it.
   */
  ask(question: string, context: string, onChunk: (text: string) => void, signal: AbortSignal): Promise<void>;
}
