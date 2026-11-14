/** English UI texts of the AI summaries settings (#75). */
export const en = {
  'settings.section.summary': 'AI summaries',
  'summary.hint':
    'Summarise notes and feed articles with a local Ollama. Nothing leaves your home network; until it is set up, the feature stays hidden.',
  'summary.url': 'Ollama address (home network)',
  'summary.load_models': 'Load models',
  'summary.model': 'Model',
  'summary.models_found': 'Models found: {n}.',
  'summary.no_models': 'No models installed; pull one first, e.g. ollama pull llama3.2:3b.',
  'summary.max_input': 'Summarise at most {n} characters',
  'summary.save': 'Save',
  'summary.saved': 'Connected and saved. Notes and feed items now show a Summarise button.',
  'summary.disable': 'Switch off',
  'summary.disabled': 'Switched off; cached summaries deleted.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
