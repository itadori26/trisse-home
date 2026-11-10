import { describe, expect, it } from 'vitest';
import { excerpt, parseNotesConfig } from './note';

describe('notes domain', () => {
  it('strips markdown for excerpts', () => {
    expect(excerpt('# Title\n\n- **bold** [link](https://x)')).toBe('Title bold link https://x');
    expect(excerpt('a'.repeat(200), 10)).toHaveLength(10);
  });
  it('defaults the scope to pinned notes', () => {
    expect(parseNotesConfig({}).scope).toBe('pinned');
    expect(parseNotesConfig({ scope: 'all' }).scope).toBe('all');
  });
});
