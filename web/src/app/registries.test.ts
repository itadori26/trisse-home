import { describe, expect, it } from 'vitest';
import { placeholders } from '../shared/i18n/format';
import { en as sharedEn } from '../shared/i18n/locales/en';
import { messages as featureEn } from './messages.en';
import { messages as featureDe } from './messages.de';
import { registeredSections } from './settingsSections';

/** Files of every feature registry, keyed by path (#193). */
const serviceFiles = import.meta.glob<{ service: { key: string } }>('./services/*.ts', { eager: true });
const featureEnFiles = import.meta.glob<{ en: Record<string, string> }>('../features/*/i18n/en.ts', {
  eager: true,
});

describe('feature registries (#193)', () => {
  it('registers every service once, under the name of its file', () => {
    const keys = Object.entries(serviceFiles).map(([path, m]) => [path, m.service.key]);
    for (const [path, key] of keys) expect(path, String(key)).toBe(`./services/${key}.ts`);
    expect(new Set(keys.map(([, k]) => k)).size).toBe(keys.length);
  });

  it('has unique settings section ids that exist in the catalogue', () => {
    const ids = registeredSections.map((s) => s.id);
    expect(new Set(ids).size).toBe(ids.length);
    for (const s of registeredSections) expect(featureEn[s.label], s.id).toBeTruthy();
  });

  it('translates every feature text into German with the same placeholders', () => {
    expect(Object.keys(featureDe).sort()).toEqual(Object.keys(featureEn).sort());
    for (const [key, text] of Object.entries(featureEn)) {
      expect(placeholders(featureDe[key] ?? '').sort(), key).toEqual(placeholders(text).sort());
      expect((featureDe[key] ?? '').trim(), key).not.toBe('');
    }
  });

  it('never reuses a key of the shared catalogue or of another feature', () => {
    const seen = new Map<string, string>();
    for (const [path, m] of Object.entries(featureEnFiles)) {
      for (const key of Object.keys(m.en)) {
        expect(key in sharedEn, `${key} is also in the shared catalogue`).toBe(false);
        expect(seen.get(key), `${key} is defined by ${seen.get(key)} and ${path}`).toBeUndefined();
        seen.set(key, path);
      }
    }
  });
});
