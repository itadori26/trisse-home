import { describe, expect, it } from 'vitest';
import { bookmarklet, defaultTarget, initialForm, parseShared } from './capture';

describe('quick capture prefill (#66)', () => {
  it('reads url, title and text', () => {
    expect(parseShared('?url=https%3A%2F%2Fexample.org%2Fa&title=Example&text=')).toEqual({
      url: 'https://example.org/a',
      title: 'Example',
      text: '',
    });
  });

  it('takes a shared link out of the text (Android share)', () => {
    expect(
      parseShared('?title=&text=' + encodeURIComponent('Look at this https://example.org/x?y=1 now')),
    ).toEqual({
      url: 'https://example.org/x?y=1',
      title: '',
      text: 'Look at this now',
    });
  });

  it('drops links that are not http(s) and limits the length', () => {
    expect(parseShared('?url=javascript%3Aalert(1)').url).toBe('');
    expect(parseShared('?title=' + 'x'.repeat(500)).title).toHaveLength(200);
  });

  it('suggests a bookmark for links and a note for text', () => {
    expect(defaultTarget({ url: 'https://a.b/', title: '', text: '' })).toBe('bookmark');
    expect(defaultTarget({ url: '', title: '', text: 'Idee' })).toBe('note');
    expect(initialForm({ url: 'https://a.b/', title: 'A', text: 'quote' })).toEqual({
      target: 'bookmark',
      url: 'https://a.b/',
      title: 'A',
      text: '',
    });
    expect(initialForm({ url: '', title: '', text: 'Idee: Fahrradtour' })).toEqual({
      target: 'note',
      url: '',
      title: '',
      text: 'Idee: Fahrradtour',
    });
  });

  it('builds a bookmarklet for this server', () => {
    const code = bookmarklet('http://home:8090');
    expect(code.split(':')[0]).toBe('javascript'); // a bookmarklet is a script link by design
    expect(code).toContain("'http://home:8090/capture?url='");
    expect(code).toContain('getSelection()');
  });
});
