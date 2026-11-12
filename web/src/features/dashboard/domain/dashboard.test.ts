import { describe, expect, it } from 'vitest';
import {
  initialTabId,
  isCidr,
  parseNetworks,
  reorder,
  type TabRule,
  validateBackground,
  validateRules,
  validateTabName,
} from './dashboard';

describe('validateTabName', () => {
  it.each([
    ['', /empty/],
    ['   ', /empty/],
    ['x'.repeat(41), /40/],
  ])('rejects %j', (name, msg) => expect(validateTabName(name)).toMatch(msg));
  it('accepts normal names', () => expect(validateTabName(' Server ')).toBeNull());
});

describe('reorder', () => {
  it('moves elements', () => {
    expect(reorder(['a', 'b', 'c'], 0, 2)).toEqual(['b', 'c', 'a']);
    expect(reorder(['a', 'b', 'c'], 2, 0)).toEqual(['c', 'a', 'b']);
  });

  it('validates tab backgrounds', () => {
    expect(validateBackground('')).toBeNull();
    expect(validateBackground('https://img.example/a.jpg')).toBeNull();
    expect(validateBackground('javascript:alert(1)')).toMatch(/http/);
  });
});

describe('tab rules (#71)', () => {
  const rule = (r: Partial<TabRule>): TabRule => ({ days: [], from: '', to: '', networks: [], ...r });

  it('accepts complete rules', () => {
    expect(
      validateRules([
        rule({ days: [1, 5], from: '22:00', to: '02:00' }),
        rule({ networks: ['192.168.178.0/24', '10.13.13.0/24', 'fd00::/8'] }),
        rule({}),
      ]),
    ).toBeNull();
  });

  it.each([
    [rule({ from: '07:00' }), /both/],
    [rule({ from: '7:00', to: '09:00' }), /07:30/],
    [rule({ from: '24:00', to: '09:00' }), /07:30/],
    [rule({ networks: ['lan'] }), /lan/],
    [rule({ networks: ['192.168.300.0/24'] }), /300/],
    [rule({ networks: ['10.0.0.0/33'] }), /33/],
    [rule({ networks: Array.from({ length: 11 }, (_, i) => `10.0.${i}.0/24`) }), /10/],
  ])('rejects %j', (r, msg) => expect(validateRules([r])).toMatch(msg));

  it('limits the number of rules', () => {
    expect(validateRules(Array.from({ length: 11 }, () => rule({})))).toMatch(/10/);
  });

  it('parses network lists', () => {
    expect(parseNetworks(' 10.0.0.0/8, 192.168.178.0/24\nfd00::/8 ,')).toEqual([
      '10.0.0.0/8',
      '192.168.178.0/24',
      'fd00::/8',
    ]);
    expect(isCidr('10.0.0.1')).toBe(false);
  });

  it('picks the initial tab: manual choice, rule, last used, first', () => {
    const ids = ['home', 'eve', 'srv'];
    expect(initialTabId(ids, { manual: 'srv', byRule: 'eve', lastUsed: 'home' })).toBe('srv');
    expect(initialTabId(ids, { manual: '', byRule: 'eve', lastUsed: 'srv' })).toBe('eve');
    expect(initialTabId(ids, { manual: 'gone', byRule: '', lastUsed: 'srv' })).toBe('srv');
    expect(initialTabId(ids, { manual: '', byRule: 'gone', lastUsed: '' })).toBe('home');
    expect(initialTabId([], { manual: '', byRule: '', lastUsed: '' })).toBe('');
  });
});
