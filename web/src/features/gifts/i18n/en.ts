/** English UI texts of the gift ideas feature (#684). */
export const en = {
  'widget.gifts.title': 'Gift ideas',
  'widget.gifts.description': 'Private gift ideas per person, next to their birthday.',
  'compact.example.gifts': '3 gift ideas',
  'gifts.new': 'New idea',
  'gifts.edit': 'Change idea',
  'gifts.private_hint': 'Only you can see your gift ideas.',
  'gifts.person': 'For',
  'gifts.idea': 'Idea',
  'gifts.link': 'Link',
  'gifts.note': 'Note',
  'gifts.bought': 'Bought',
  'gifts.empty': 'No gift ideas yet. Only you will see them.',
  'gifts.birthday_today': 'Birthday today',
  'gifts.birthday_in': 'Birthday {date}, in {n} days',
  'gifts.add_for': 'New idea for {person}',
  'gifts.bought_named': 'Bought: {title}',
  'gifts.edit_named': 'Change {title}',
  'gifts.remove': 'Remove {title}',
  'gifts.error.person': 'Enter a name of up to 60 characters.',
  'gifts.error.title': 'Enter an idea of up to 120 characters.',
  'gifts.error.link': 'The link must start with http:// or https://.',
  'gifts.error.note': 'The note may be at most 500 characters long.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
