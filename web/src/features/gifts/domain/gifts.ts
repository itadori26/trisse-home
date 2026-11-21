/** A private gift idea (#684). */
export interface GiftIdea {
  id: string;
  person: string;
  title: string;
  link: string;
  note: string;
  bought: boolean;
}

/** What the dialog edits. */
export type GiftInput = Omit<GiftIdea, 'id'>;

/** The next birthday of a person, from the birthdays module. */
export interface NextBirthday {
  name: string;
  /** YYYY-MM-DD */
  next: string;
  inDays: number;
}

/** The ideas for one person with their next birthday, if known. */
export interface PersonGroup {
  person: string;
  birthday: NextBirthday | null;
  ideas: GiftIdea[];
}

/** Limits; they mirror the backend. */
export const MAX_PERSON_LENGTH = 60;
export const MAX_TITLE_LENGTH = 120;
export const MAX_NOTE_LENGTH = 500;

const key = (name: string) => name.trim().replace(/\s+/g, ' ').toLowerCase();

/**
 * Groups ideas by person, matching names without case. People with a known
 * birthday come first, the nearest one at the top; the rest follow by name.
 */
export function groupByPerson(ideas: GiftIdea[], birthdays: NextBirthday[]): PersonGroup[] {
  const byName = new Map(birthdays.map((b) => [key(b.name), b]));
  const groups = new Map<string, PersonGroup>();
  for (const idea of ideas) {
    const k = key(idea.person);
    let g = groups.get(k);
    if (!g) {
      g = { person: idea.person, birthday: byName.get(k) ?? null, ideas: [] };
      groups.set(k, g);
    }
    g.ideas.push(idea);
  }
  return [...groups.values()].sort((a, b) => {
    if (a.birthday && b.birthday) return a.birthday.inDays - b.birthday.inDays;
    if (a.birthday || b.birthday) return a.birthday ? -1 : 1;
    return a.person.localeCompare(b.person);
  });
}

/** Message keys of validateGift. */
export type GiftError = 'gifts.error.person' | 'gifts.error.title' | 'gifts.error.link' | 'gifts.error.note';

/** Validates the dialog and returns a message key or null. */
export function validateGift(input: GiftInput): GiftError | null {
  const person = input.person.trim();
  if (!person || [...person].length > MAX_PERSON_LENGTH) return 'gifts.error.person';
  const title = input.title.trim();
  if (!title || [...title].length > MAX_TITLE_LENGTH) return 'gifts.error.title';
  const link = input.link.trim();
  if (link && !/^https?:\/\/[^\s/]+/i.test(link)) return 'gifts.error.link';
  if ([...input.note.trim()].length > MAX_NOTE_LENGTH) return 'gifts.error.note';
  return null;
}
