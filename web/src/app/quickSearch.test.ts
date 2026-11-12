import { describe, expect, it } from 'vitest';
import { fold, type QuickItem, score, search } from './quickSearch';

const items: QuickItem[] = [
  {
    kind: 'bookmark',
    id: '1',
    title: 'Go documentation',
    detail: 'https://go.dev/doc',
    keywords: ['golang'],
  },
  { kind: 'service', id: '2', title: 'Nextcloud', detail: 'https://cloud.example' },
  { kind: 'note', id: '3', title: 'Router setup', detail: 'Fritz!Box notes' },
  { kind: 'todo', id: '4', title: 'Büro aufräumen', detail: '' },
  { kind: 'bookmark', id: '5', title: 'Nginx', detail: 'https://nginx.org' },
];

describe('quick search', () => {
  it('ranks title prefixes first and groups by kind', () => {
    const groups = search(items, 'n');
    expect(groups.map((g) => g.kind)).toEqual(['bookmark', 'service', 'note', 'todo']);
    expect(groups[0]?.items[0]?.title).toBe('Nginx');
  });
  it('matches every word, keywords and ignores diacritics', () => {
    const go = items[0] as QuickItem;
    expect(score(go, 'golang doc')).toBeGreaterThan(0);
    expect(score(go, 'golang rust')).toBe(0);
    expect(search(items, 'buro')[0]?.items[0]?.id).toBe('4');
    expect(fold('Ärger')).toBe('arger');
  });
  it('returns nothing for blank queries', () => {
    expect(search(items, '  ')).toEqual([]);
  });
});
