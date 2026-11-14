/** English UI texts added to the rss feature (#125, #193). */
export const en = {
  'rss.compact': 'Compact (titles only)',
  'rss.opml_import': 'Import OPML…',
  'rss.opml_export': 'Export OPML',
  'rss.opml_report': '{added} feeds added, {duplicates} already subscribed, {invalid} not imported.',
  'rss.opml_too_large': 'The OPML file is larger than 2 MB.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
