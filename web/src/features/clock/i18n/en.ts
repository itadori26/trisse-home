/** English UI texts added to the clock feature (#122, #193). */
export const en = {
  'clock.show_week': 'Show calendar week',
  'clock.week': 'Week {n}',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
