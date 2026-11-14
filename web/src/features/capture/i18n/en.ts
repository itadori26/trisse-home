/** English UI texts of the quick capture page (#66). */
export const en = {
  'capture.title': 'Save to trisse-home',
  'capture.as': 'Save as',
  'capture.bookmark': 'Bookmark',
  'capture.todo': 'To-do',
  'capture.note': 'Note',
  'capture.name': 'Title',
  'capture.url': 'Link',
  'capture.text': 'Text',
  'capture.save': 'Save',
  'capture.missing': 'Fill in a title (or a valid link for a bookmark).',
  'capture.saved_bookmark': 'Bookmark saved.',
  'capture.saved_todo': 'To-do saved.',
  'capture.saved_note': 'Note saved.',
  'capture.undo': 'Undo',
  'capture.undone': 'Removed again.',
  'capture.close': 'Close',
  'capture.dashboard': 'Open dashboard',
  'capture.another': 'Save something else',
  'capture.bookmarklet': 'Quick capture from any page',
  'capture.bookmarklet_hint':
    'Drag this link to your bookmarks bar. On any page it opens a small window that saves the page as a bookmark, to-do or note (a selected text is taken along). On phones, share to the installed app instead.',
  'capture.bookmarklet_link': '＋ trisse-home',
  'capture.bookmarklet_copy': 'Copy bookmarklet',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
