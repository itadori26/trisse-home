import { describe, expect, it } from 'vitest';
import { boundedDistance, fold, matchToken, maxDistance, tokenize } from './text';

describe('tokenize', () => {
  it('folds, splits and drops stop words', () => {
    expect(tokenize('Wie voll ist die Platte?')).toEqual(['voll', 'platte']);
    expect(tokenize('Wetter in Köln morgen')).toEqual(['wetter', 'koln', 'morgen']);
    expect(tokenize('restart my-app.local!')).toEqual(['restart', 'my-app.local']);
    expect(tokenize('  ')).toEqual([]);
  });
  it('keeps umlauts comparable', () => {
    expect(fold('Größe')).toBe('grosse');
  });
});

describe('boundedDistance', () => {
  it('counts transpositions as one edit', () => {
    expect(boundedDistance('wtter', 'wetter', 2)).toBe(1);
    expect(boundedDistance('wteter', 'wetter', 2)).toBe(1);
    expect(boundedDistance('jellyifn', 'jellyfin', 2)).toBe(1);
  });
  it('stops early above the limit', () => {
    expect(boundedDistance('abcdefgh', 'zyxwvuts', 1)).toBe(2);
    expect(boundedDistance('a', 'abcdef', 2)).toBe(3);
  });
});

describe('matchToken', () => {
  const cases: [string, string, string | null][] = [
    // exact and prefix
    ['wetter', 'wetter', 'exact'],
    ['we', 'wetter', 'prefix'],
    ['jel', 'jellyfin', 'prefix'],
    ['w', 'wetter', null], // prefix needs two characters
    // typos within the bound
    ['wetetr', 'wetter', 'fuzzy'],
    ['jellyfn', 'jellyfin', 'fuzzy'],
    ['nextclodu', 'nextcloud', 'fuzzy'],
    ['platet', 'platte', 'fuzzy'],
    ['speicher', 'speicher', 'exact'],
    // umlauts after fold
    [fold('Größe'), fold('grösse'), 'exact'],
    // false positives that must not match
    ['cpu', 'cup', null], // short tokens: exact or prefix only
    ['ram', 'rom', null],
    ['disk', 'dash', null], // two edits on a short term
    ['docker', 'locker', 'fuzzy'], // one substitution is allowed from 4 characters …
    ['dockerx', 'nginx', null], // … but not a different word
    ['wetter', 'weather', null], // distance 2 on 7 characters is too far
  ];
  it.each(cases)('%s ~ %s → %s', (token, term, want) => {
    expect(matchToken(token, term)).toBe(want);
  });
  it('bounds the distance by length', () => {
    expect([3, 4, 7, 8].map(maxDistance)).toEqual([0, 1, 1, 2]);
  });
});

describe('performance (#157)', () => {
  it('matches 500 names in well under 2 ms per keystroke', () => {
    const names = Array.from({ length: 500 }, (_, i) => `service-${i.toString(36)}-container`);
    const tokens = ['jellyfn', 'wetter', 'serv', 'nextclod', 'cpu'];
    const start = performance.now();
    const rounds = 20;
    for (let r = 0; r < rounds; r++) for (const t of tokens) for (const n of names) matchToken(t, n);
    const perKeystroke = (performance.now() - start) / rounds;
    // CI machines vary; a desktop needs ~0.2 ms here, a mid-range phone ~5× more.
    expect(perKeystroke).toBeLessThan(2);
  });
});
