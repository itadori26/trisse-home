import { describe, expect, it } from 'vitest';
import { countText, excerpt, parseNotesConfig } from './note';

describe('countText (#121)', () => {
  it('counts nothing for an empty or blank note', () => {
    expect(countText('')).toEqual({ words: 0, characters: 0 });
    // Only the line break is not counted: two spaces, a space, a tab and a space remain.
    expect(countText('  \n \t ')).toEqual({ words: 0, characters: 5 });
  });

  it('ignores repeated spaces and line breaks between words', () => {
    expect(countText('one   two\n\nthree\t four').words).toBe(4);
    expect(countText('a b').characters).toBe(3);
    expect(countText('a\nb').characters).toBe(2);
  });

  it('does not count Markdown syntax as words', () => {
    expect(countText('# Title').words).toBe(1);
    expect(countText('- a\n- b\n---\n> c').words).toBe(3);
    const md = '# Title\n\n- item one\n\n---\n\n> quoted *text* here\n\n```js\nconst x = 1;\n```\n';
    // Title (1) + item one (2) + quoted text here (3) + const x 1 (3). The fence
    // lines themselves and the symbols #, -, ---, >, =, ; are not words.
    expect(countText(md).words).toBe(9);
  });

  it('counts only the text of links and images', () => {
    expect(countText('see [the docs](https://example.com/very/long/path) now').words).toBe(4);
    expect(countText('![diagram of the setup](img.png)').words).toBe(4);
  });

  it('counts letters with accents and digits as words', () => {
    expect(countText('Über 3 Brücken müssen wir gehn').words).toBe(6);
    expect(countText('2026-11-13').words).toBe(1);
  });

  it('counts an emoji or a letter with an accent as one character', () => {
    expect(countText('👍').characters).toBe(1);
    expect(countText('👨‍👩‍👧 family').characters).toBe(8);
    expect(countText('🇩🇪').characters).toBe(1);
    expect(countText('é').characters).toBe(1);
  });
});

describe('notes domain', () => {
  it('strips markdown for excerpts', () => {
    expect(excerpt('# Title\n\n- **bold** [link](https://x)')).toBe('Title bold link https://x');
    expect(excerpt('a'.repeat(200), 10)).toHaveLength(10);
  });
  it('defaults the scope to pinned notes', () => {
    expect(parseNotesConfig({}).scope).toBe('pinned');
    expect(parseNotesConfig({ scope: 'all' }).scope).toBe('all');
  });
});
