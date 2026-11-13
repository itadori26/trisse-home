/** English UI texts of the launcher feature (#193). */
export const en = {
  'launcher.title': 'Services and bookmarks',
  'launcher.filter': 'Filter services and bookmarks…',
  'launcher.show': 'Show',
  'launcher.all': 'All',
  'launcher.services': 'Services',
  'launcher.bookmarks': 'Bookmarks',
  'launcher.show_all_services': 'Show all {n}',
  'launcher.show_all_bookmarks': 'Show all {n}',
  'launcher.no_services': 'No matching services.',
  'launcher.no_bookmarks': 'No matching bookmarks.',
  'launcher.online_ms': 'online · {ms} ms',
  'launcher.offline': 'offline',
  'launcher.unknown': 'not checked yet',
  'launcher.footer_before': 'Choose what appears here in',
  'launcher.footer_link': 'Settings → Appearance',
  'launcher.footer_ctrl_k': 'Everything is also searchable with Ctrl K',
  'launcher.all_online': 'All services online',
  'launcher.some_offline': '{n} offline',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
