/** English UI texts added to the bookmarks feature (#120, #193). */
export const en = {
  'bookmarks.sort': 'Sort order',
  'bookmarks.sort_manual': 'Default',
  'bookmarks.sort_az': 'A–Z',
  'bookmarks.sort_newest': 'Newest first',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
