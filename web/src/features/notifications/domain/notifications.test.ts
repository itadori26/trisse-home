import { describe, expect, it } from 'vitest';
import { suggestTopic } from './notifications';

describe('suggestTopic', () => {
  it('creates a long, valid topic', () => {
    const t = suggestTopic(() => 0.5);
    expect(t).toMatch(/^trisse-home-[a-z0-9]{12}$/);
  });
});
