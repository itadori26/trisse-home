/** English UI texts added to the notes feature (#121, #193). */
export const en = {
  'notes.word_count_one': '{n} word',
  'notes.word_count_other': '{n} words',
  'notes.char_count_one': '{n} character',
  'notes.char_count_other': '{n} characters',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
