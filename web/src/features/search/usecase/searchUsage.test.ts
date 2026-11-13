import { describe, expect, it } from 'vitest';
import { LocalUsageStore } from '../adapter/localUsageStore';
import { SearchUsage, type UsageState, type UsageStore } from './searchUsage';

class MemoryStore implements UsageStore {
  saved: UsageState = { enabled: true, entries: [] };
  load() {
    return this.saved;
  }
  save(s: UsageState) {
    this.saved = s;
  }
}

describe('SearchUsage (#159)', () => {
  it('records choices and boosts them, persisting through the store', () => {
    let now = Date.UTC(2026, 10, 13);
    const store = new MemoryStore();
    const usage = new SearchUsage(store, () => now);
    usage.record('plex', 'service:1');
    now += 1000;
    expect(usage.boost('pl', 'service:1')).toBeCloseTo(1);
    expect(usage.score('pl', 'service:1', 2)).toBeCloseTo(3);
    expect(usage.score('pl', 'bookmark:2', 2)).toBe(2);
    expect(new SearchUsage(store, () => now).size).toBe(1);
  });

  it('switched off it neither records nor boosts, and forgets', () => {
    const store = new MemoryStore();
    const usage = new SearchUsage(store);
    usage.record('plex', 'service:1');
    usage.setEnabled(false);
    expect(usage.size).toBe(0);
    usage.record('plex', 'service:1');
    expect(usage.boost('plex', 'service:1')).toBe(0);
    expect(store.saved).toEqual({ enabled: false, entries: [] });
  });

  it('clears the history', () => {
    const usage = new SearchUsage(new MemoryStore());
    usage.record('plex', 'service:1');
    usage.clear();
    expect(usage.size).toBe(0);
    expect(usage.enabled).toBe(true);
  });
});

describe('LocalUsageStore', () => {
  const fake = () => {
    const data = new Map<string, string>();
    return {
      data,
      storage: {
        getItem: (k: string) => data.get(k) ?? null,
        setItem: (k: string, v: string) => void data.set(k, v),
      } as unknown as Storage,
    };
  };

  it('round-trips the state', () => {
    const { storage } = fake();
    const store = new LocalUsageStore(() => storage);
    store.save({ enabled: false, entries: [{ key: 'a', target: 't', value: 1, at: 1 }] });
    expect(store.load()).toEqual({ enabled: false, entries: [{ key: 'a', target: 't', value: 1, at: 1 }] });
  });

  it('works without storage and with garbage', () => {
    expect(new LocalUsageStore(() => undefined).load()).toEqual({ enabled: true, entries: [] });
    const { storage, data } = fake();
    data.set('trisse-home.search-usage', '{not json');
    expect(new LocalUsageStore(() => storage).load()).toEqual({ enabled: true, entries: [] });
    const throwing = {
      getItem: () => {
        throw new Error('blocked');
      },
      setItem: () => {
        throw new Error('full');
      },
    } as unknown as Storage;
    const store = new LocalUsageStore(() => throwing);
    expect(store.load().entries).toEqual([]);
    expect(() => store.save({ enabled: true, entries: [] })).not.toThrow();
  });
});
