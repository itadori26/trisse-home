import { describe, expect, it } from 'vitest';
import { findByTitle, linkWikiLinks, noteLinkTitle, plain } from './note';

describe('wiki links (#73)', () => {
  it('turns [[Title]] into internal links outside code', () => {
    const body = 'See [[Shopping list]] and [[ Trip 2027 | the trip]].\n`[[code]]`\n```\n[[block]]\n```';
    expect(linkWikiLinks(body)).toBe(
      'See [Shopping list](#note:Shopping%20list) and [the trip](#note:Trip%202027).\n`[[code]]`\n```\n[[block]]\n```',
    );
  });

  it('reads the title back from a link', () => {
    expect(noteLinkTitle('#note:Trip%202027')).toBe('Trip 2027');
    expect(noteLinkTitle('https://example.org')).toBeNull();
    expect(noteLinkTitle('#note:%E0%A4%A')).toBeNull();
  });

  it('resolves titles ignoring case and spaces', () => {
    const notes = [{ id: '1', title: 'Shopping list' }];
    expect(findByTitle(notes, ' shopping LIST ')?.id).toBe('1');
    expect(findByTitle(notes, 'other')).toBeUndefined();
  });

  it('joins highlighted segments', () => {
    expect(
      plain([
        { text: 'a ', hit: false },
        { text: 'b', hit: true },
      ]),
    ).toBe('a b');
  });
});
