import { describe, expect, it } from 'vitest';
import { captureCommand } from './capture';

// Thursday, 12 November 2026
const today = new Date(2026, 10, 12, 9, 30);

describe('captureCommand (#66)', () => {
  it('saves a bookmark with the host as fallback title', () => {
    expect(captureCommand({ target: 'bookmark', url: 'example.org/a', title: '', text: '' }, today)).toEqual({
      kind: 'bookmark',
      url: 'https://example.org/a',
      title: 'example.org',
    });
    expect(captureCommand({ target: 'bookmark', url: 'not a url', title: 'x', text: '' }, today)).toBeNull();
  });

  it('takes the due date of a to-do from its title', () => {
    expect(
      captureCommand({ target: 'todo', url: '', title: 'Artikel lesen morgen', text: '' }, today),
    ).toEqual({
      kind: 'todo',
      title: 'Artikel lesen',
      due: '2026-11-13',
    });
    expect(captureCommand({ target: 'todo', url: '', title: ' ', text: '' }, today)).toBeNull();
  });

  it('keeps title, text and link in a note', () => {
    expect(
      captureCommand({ target: 'note', url: 'https://a.b/', title: 'Idee', text: 'Fahrradtour' }, today),
    ).toEqual({
      kind: 'note',
      title: 'Idee\nFahrradtour\nhttps://a.b/',
    });
    expect(captureCommand({ target: 'note', url: '', title: '', text: '' }, today)).toBeNull();
  });
});
