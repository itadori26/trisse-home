/** English UI texts of the read-later feature (#99). */
export const en = {
  'widget.readlater.title': 'Read later',
  'widget.readlater.description':
    'Saved articles with a clean reader copy that stays readable when the page is gone.',
  'settings.section.readlater': 'Read later',
  'readlater.pending': 'Reader copy follows in the background',
  'readlater.more': 'Cleanup and import',
  'readlater.archive_days': 'Delete archived articles',
  'readlater.after_days': 'after {n} days',
  'readlater.keep': 'never',
  'readlater.import_pocket': 'Import Pocket export…',
  'readlater.import_hint':
    'The CSV file from the Pocket export. The links are added at once; reader copies follow in the background, a few every five minutes.',
  'readlater.import_report': '{added} added, {duplicates} already saved, {invalid} not importable.',
  'readlater.import_too_large': 'The export file must be at most 5 MB.',
  'readlater.read_later': 'Read later',
  'readlater.read_later_item': 'Read {title} later',
  'readlater.saved_for_later': 'Saved for later',
  'readlater.intro':
    'Saved pages get a clean reader copy with their images, stored on this server. Requests go only to the page you save, once.',
  'readlater.save': 'Save',
  'readlater.save_label': 'Address of the page to read later',
  'readlater.save_placeholder': '+ Read later: paste a link',
  'readlater.saving': 'Saving the page…',
  'readlater.url_invalid': 'Enter the http(s) address of a web page.',
  'readlater.empty': 'Nothing to read.',
  'readlater.minutes': '{n} min',
  'readlater.saved_on': 'saved {date}',
  'readlater.original': 'Original',
  'readlater.no_copy': 'There is no reader copy of this page.',
  'readlater.archive': 'Archive',
  'readlater.mark_read': 'Mark “{title}” as read',
  'readlater.archive_item': 'Archive “{title}”',
  'readlater.restore_item': 'Back to unread: “{title}”',
  'readlater.delete_item': 'Delete “{title}”',
  'readlater.delete_confirm': 'Delete “{title}” with its reader copy?',
  'readlater.status': 'Show',
  'readlater.status_unread': 'Unread',
  'readlater.status_read': 'Read',
  'readlater.status_archived': 'Archived',
  'readlater.storage': 'These articles use {size}.',
  'readlater.count': 'Articles: {n}',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
