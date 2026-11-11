import { describe, expect, it } from 'vitest';
import { suggestListUrl, withDone } from './shopping';

describe('shopping domain', () => {
  it('suggests a list next to the calendar', () => {
    expect(suggestListUrl('https://c/remote.php/dav/calendars/me/personal/')).toBe(
      'https://c/remote.php/dav/calendars/me/shopping/',
    );
    expect(suggestListUrl('nonsense')).toBe('');
  });
  it('updates done immutably', () => {
    const items = [{ id: 'a', summary: 'Milk', done: false }];
    const next = withDone(items, 'a', true);
    expect(next[0]?.done).toBe(true);
    expect(items[0]?.done).toBe(false);
  });
});
