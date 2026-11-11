import { describe, expect, it } from 'vitest';
import { formatMessage, placeholders, resolveLocale } from './format';
import { de } from './locales/de';
import { en } from './locales/en';

describe('i18n', () => {
  it('formats placeholders', () => {
    expect(formatMessage('Delete {name}?', { name: 'Home' })).toBe('Delete Home?');
    expect(formatMessage('{a} of {b}', { a: 1 })).toBe('1 of {b}');
  });

  it('resolves the locale', () => {
    expect(resolveLocale('', ['de-AT', 'en'])).toBe('de');
    expect(resolveLocale('en', ['de-DE'])).toBe('en');
    expect(resolveLocale('', ['fr-FR'])).toBe('en');
  });

  it('has a German text for every key with the same placeholders', () => {
    const missing = Object.keys(en).filter((k) => !(k in de));
    expect(missing).toEqual([]);
    const extra = Object.keys(de).filter((k) => !(k in en));
    expect(extra).toEqual([]);
    for (const [key, text] of Object.entries(en) as [keyof typeof en, string][]) {
      expect(placeholders(de[key]), key).toEqual(placeholders(text));
      expect(de[key].trim(), key).not.toBe('');
    }
  });
});
