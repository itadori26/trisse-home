import type { en } from './en';

/** German UI texts of the gift ideas feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.gifts.title': 'Geschenkideen',
  'widget.gifts.description': 'Private Geschenkideen pro Person, neben ihrem Geburtstag.',
  'compact.example.gifts': '3 Geschenkideen',
  'gifts.new': 'Neue Idee',
  'gifts.edit': 'Idee ändern',
  'gifts.private_hint': 'Deine Geschenkideen siehst nur du.',
  'gifts.person': 'Für',
  'gifts.idea': 'Idee',
  'gifts.link': 'Link',
  'gifts.note': 'Notiz',
  'gifts.bought': 'Gekauft',
  'gifts.empty': 'Noch keine Geschenkideen. Nur du wirst sie sehen.',
  'gifts.birthday_today': 'Heute Geburtstag',
  'gifts.birthday_in': 'Geburtstag am {date}, in {n} Tagen',
  'gifts.add_for': 'Neue Idee für {person}',
  'gifts.bought_named': 'Gekauft: {title}',
  'gifts.edit_named': '{title} ändern',
  'gifts.remove': '{title} entfernen',
  'gifts.error.person': 'Bitte einen Namen mit bis zu 60 Zeichen eingeben.',
  'gifts.error.title': 'Bitte eine Idee mit bis zu 120 Zeichen eingeben.',
  'gifts.error.link': 'Der Link muss mit http:// oder https:// beginnen.',
  'gifts.error.note': 'Die Notiz darf höchstens 500 Zeichen lang sein.',
};
