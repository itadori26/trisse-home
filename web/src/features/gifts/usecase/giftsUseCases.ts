import type { GiftIdea, GiftInput, NextBirthday } from '../domain/gifts';

/** Output port: gift idea backend. */
export interface GiftsGateway {
  list(): Promise<GiftIdea[]>;
  create(input: GiftInput): Promise<GiftIdea>;
  update(id: string, input: GiftInput): Promise<GiftIdea>;
  remove(id: string): Promise<void>;
}

/**
 * Output port: the next birthdays of the birthdays module. The app layer
 * implements it and answers an empty list when that module is off.
 */
export interface BirthdaySource {
  upcoming(): Promise<NextBirthday[]>;
}

const clean = (i: GiftInput): GiftInput => ({
  person: i.person.trim().replace(/\s+/g, ' '),
  title: i.title.trim(),
  link: i.link.trim(),
  note: i.note.trim(),
  bought: i.bought,
});

/** Application logic of the gift ideas (#684). */
export class GiftsUseCases {
  constructor(private readonly gateway: GiftsGateway) {}

  list(): Promise<GiftIdea[]> {
    return this.gateway.list();
  }

  /** Creates (no id) or changes an idea. */
  save(id: string | null, input: GiftInput): Promise<GiftIdea> {
    return id ? this.gateway.update(id, clean(input)) : this.gateway.create(clean(input));
  }

  /** Marks an idea bought or open again. */
  setBought(idea: GiftIdea, bought: boolean): Promise<GiftIdea> {
    const { id, ...input } = idea;
    return this.gateway.update(id, { ...input, bought });
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }
}
