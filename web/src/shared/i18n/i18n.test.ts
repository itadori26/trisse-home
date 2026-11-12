import { describe, expect, it } from 'vitest';
import { formatMessage, placeholders, resolveLocale } from './format';
import { de } from './locales/de';
import { en } from './locales/en';
import { serverDe } from './locales/server-de';
import { serverMessages } from '../api-client/generated/server-messages';
import { serverMessage, setLanguage } from './i18n.svelte';

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

describe('backend messages (#102)', () => {
  it('has a German translation with the same placeholders for every backend message', () => {
    for (const [key, english] of Object.entries(serverMessages)) {
      const german = serverDe[key as keyof typeof serverDe];
      expect(german, key).toBeTruthy();
      expect(placeholders(german).sort(), english).toEqual(placeholders(english).sort());
    }
  });

  it('translates known keys and falls back for unknown ones', async () => {
    const key = Object.keys(serverMessages).find((k) =>
      serverMessages[k as keyof typeof serverMessages].includes('{max}'),
    );
    expect(key).toBeDefined();
    setLanguage('de');
    expect(serverMessage(key, { max: '500' })).toContain('500');
    expect(serverMessage('srv_unknown')).toBeUndefined();
    expect(serverMessage(undefined)).toBeUndefined();
    setLanguage('en');
  });
});
