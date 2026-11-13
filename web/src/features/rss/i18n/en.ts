/** English UI texts added to the rss feature (#125, #193). */
export const en = {
  'rss.compact': 'Compact (titles only)',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
