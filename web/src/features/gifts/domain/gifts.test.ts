import { describe, expect, it } from 'vitest';
import { groupByPerson, validateGift, type GiftIdea } from './gifts';

const idea = (id: string, person: string): GiftIdea => ({
  id,
  person,
  title: id,
  link: '',
  note: '',
  bought: false,
});

describe('gift ideas domain', () => {
  it('groups by person and puts the nearest birthday first', () => {
    const groups = groupByPerson(
      [idea('a', 'Zoe'), idea('b', 'paul'), idea('c', 'Anna'), idea('d', 'Paul ')],
      [
        { name: 'Paul', next: '2026-12-20', inDays: 29 },
        { name: 'Zoe', next: '2026-11-25', inDays: 4 },
      ],
    );
    expect(groups.map((g) => g.person)).toEqual(['Zoe', 'paul', 'Anna']);
    expect(groups[1]?.ideas.map((i) => i.id)).toEqual(['b', 'd']);
    expect(groups[2]?.birthday).toBeNull();
  });

  it('validates the dialog', () => {
    const ok = { person: 'Ben', title: 'Book', link: '', note: '', bought: false };
    expect(validateGift(ok)).toBeNull();
    expect(validateGift({ ...ok, person: ' ' })).toBe('gifts.error.person');
    expect(validateGift({ ...ok, title: '' })).toBe('gifts.error.title');
    expect(validateGift({ ...ok, link: 'ftp://x' })).toBe('gifts.error.link');
    expect(validateGift({ ...ok, note: 'x'.repeat(501) })).toBe('gifts.error.note');
  });
});
