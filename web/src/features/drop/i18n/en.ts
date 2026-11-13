/** English UI texts of the drop zone (#101). */
export const en = {
  'widget.drop.title': 'Drop zone',
  'widget.drop.description':
    'Share text, links and small files between your devices; items expire automatically.',
  'drop.zone': 'Drop zone: paste text or drop files here',
  'drop.text': 'Text or link',
  'drop.placeholder': 'Paste text or a link, or drop files here…',
  'drop.expiry': 'Keep for',
  'drop.expiry_1': '1 hour',
  'drop.expiry_24': '24 hours',
  'drop.expiry_168': '7 days',
  'drop.file': 'File…',
  'drop.drop': 'Drop',
  'drop.copy': 'Copy',
  'drop.download': 'Download {name}',
  'drop.delete': 'Delete',
  'drop.empty': 'Nothing here. Drop something to pick it up on another device.',
  'drop.expires': 'expires {when}',
  'drop.in_minutes': 'in {n} min',
  'drop.in_hours': 'in {n} h',
  'drop.in_days': 'in {n} days',
  'drop.file_empty': 'The file is empty.',
  'drop.file_too_large': 'The file must be at most 25 MB.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
