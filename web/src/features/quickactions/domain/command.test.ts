import { describe, expect, it } from 'vitest';
import {
  captureRest,
  mergePrefixes,
  normaliseUrl,
  palettePrefixes,
  parseQuick,
  prefixesFromConfig,
  searchPrefixes,
} from './command';
import { extractDue } from './dates';

// Thursday, 12 November 2026
const today = new Date(2026, 10, 12, 9, 30);

describe('date extraction (#105, #67)', () => {
  const cases: [string, string, string][] = [
    ['Müll rausbringen morgen', 'Müll rausbringen', '2026-11-13'],
    ['take out trash tomorrow', 'take out trash', '2026-11-13'],
    ['heute Steuer', 'Steuer', '2026-11-12'],
    ['Arzt übermorgen', 'Arzt', '2026-11-14'],
    ['call mum on friday', 'call mum', '2026-11-13'],
    ['Termin am Donnerstag', 'Termin', '2026-11-19'], // today is Thursday: next week
    ['Klausur 2027-02-14', 'Klausur', '2027-02-14'],
    ['Pay rent @2026-12-01', 'Pay rent', '2026-12-01'],
    ['Geburtstag 14.2.', 'Geburtstag', '2027-02-14'], // past this year: next year
    ['Party 24.12.', 'Party', '2026-12-24'],
    ['Abgabe am 3.1.2027', 'Abgabe', '2027-01-03'],
    ['Review in 3 Tagen', 'Review', '2026-11-15'],
    ['renew in 2 weeks', 'renew', '2026-11-26'],
    ['Urlaub nächste Woche planen', 'Urlaub planen', '2026-11-19'],
    ['Morgenroutine', 'Morgenroutine', ''], // no whole-word match
    ['30.2. invalid', '30.2. invalid', ''],
  ];
  for (const [input, text, due] of cases) {
    it(input, () => expect(extractDue(input, today)).toEqual({ text, due }));
  }

  it('handles month ends, midnight and the DST change', () => {
    expect(extractDue('x morgen', new Date(2026, 0, 31, 23, 59)).due).toBe('2026-02-01');
    expect(extractDue('x morgen', new Date(2026, 11, 31, 0, 0)).due).toBe('2027-01-01');
    // Europe switches to summer time on 29 March 2026.
    expect(extractDue('x morgen', new Date(2026, 2, 28, 23, 59)).due).toBe('2026-03-29');
    expect(extractDue('x in 2 days', new Date(2026, 9, 24, 23, 30)).due).toBe('2026-10-26');
  });
});

describe('quick commands', () => {
  it('parses search bar prefixes only when followed by a space', () => {
    expect(parseQuick('t Milch kaufen', searchPrefixes, today)).toEqual({
      kind: 'todo',
      title: 'Milch kaufen',
      due: '',
    });
    expect(parseQuick('t Müll morgen', searchPrefixes, today)).toEqual({
      kind: 'todo',
      title: 'Müll',
      due: '2026-11-13',
    });
    expect(parseQuick('n Idee: Fahrradtour', searchPrefixes, today)).toEqual({
      kind: 'note',
      title: 'Idee: Fahrradtour',
    });
    expect(parseQuick('b example.org Beispiel', searchPrefixes, today)).toEqual({
      kind: 'bookmark',
      url: 'https://example.org/',
      title: 'Beispiel',
    });
    expect(parseQuick('tomorrow weather', searchPrefixes, today)).toBeNull();
    expect(parseQuick('!w Berlin', searchPrefixes, today)).toBeNull();
    expect(parseQuick('t ', searchPrefixes, today)).toBeNull();
    expect(parseQuick('b not a url', searchPrefixes, today)).toBeNull();
  });

  it('parses palette commands', () => {
    expect(parseQuick('todo: Müll rausbringen morgen', palettePrefixes, today)).toEqual({
      kind: 'todo',
      title: 'Müll rausbringen',
      due: '2026-11-13',
    });
    expect(parseQuick('note:Einkauf', palettePrefixes, today)).toEqual({ kind: 'note', title: 'Einkauf' });
    expect(parseQuick('bm: https://x.org/a', palettePrefixes, today)).toEqual({
      kind: 'bookmark',
      url: 'https://x.org/a',
      title: 'x.org',
    });
    expect(parseQuick('restart nextcloud', palettePrefixes, today)).toEqual({
      kind: 'restart',
      target: 'nextcloud',
    });
    expect(parseQuick('wake NAS', palettePrefixes, today)).toEqual({ kind: 'wake', target: 'NAS' });
    expect(parseQuick('restarting', palettePrefixes, today)).toBeNull();
  });

  it('reads configured prefixes', () => {
    expect(prefixesFromConfig({ todo: 'a', note: '  ', bookmark: '!!' })).toEqual({
      todo: ['a'],
      note: ['n'],
      bookmark: ['b'],
    });
    expect(parseQuick('a Milch', prefixesFromConfig({ todo: 'a' }), today)?.kind).toBe('todo');
  });

  it('normalises URLs', () => {
    expect(normaliseUrl('example.org')).toBe('https://example.org/');
    expect(normaliseUrl('http://nas.local:5000/x')).toBe('http://nas.local:5000/x');
    expect(normaliseUrl('javascript:alert(1)')).toBeNull();
    expect(normaliseUrl('localhost')).toBeNull();
  });
});

describe('mergePrefixes (#221)', () => {
  const today = new Date(2026, 10, 13, 12);
  const both = mergePrefixes(palettePrefixes, prefixesFromConfig({ todo: 'a' }));
  it('keeps the palette prefixes and adds the search field ones', () => {
    expect(both.todo).toEqual(['todo:', 't:', 'a']);
    expect(both.note).toEqual(['note:', 'n:', 'n']);
    expect(both.restart).toEqual(['restart']);
  });
  it('lets both forms create items', () => {
    expect(parseQuick('todo: Steuer heute', both, today)?.kind).toBe('todo');
    expect(parseQuick('a Milch kaufen', both, today)?.kind).toBe('todo');
    expect(parseQuick('b example.org Beispiel', both, today)?.kind).toBe('bookmark');
  });
  it('still treats words as searches', () => {
    expect(parseQuick('apfelkuchen rezept', both, today)).toBeNull();
    expect(parseQuick('!w Oldenburg', both, today)).toBeNull();
  });
  it('does not repeat a prefix that is in both sets', () => {
    expect(mergePrefixes(searchPrefixes, searchPrefixes).todo).toEqual(['t']);
  });
});

describe('captureRest (#222)', () => {
  const both = mergePrefixes(palettePrefixes, searchPrefixes);
  it('returns the text after a capture prefix', () => {
    expect(captureRest('t Milch kaufen morgen', both)).toBe('Milch kaufen morgen');
    expect(captureRest('todo:Steuer', both)).toBe('Steuer');
    expect(captureRest('  n  Idee ', both)).toBe('Idee');
  });
  it('is null for searches and empty captures', () => {
    expect(captureRest('tomorrow', both)).toBeNull();
    expect(captureRest('t ', both)).toBeNull();
    expect(captureRest('restart nextcloud', both)).toBeNull();
  });
});
