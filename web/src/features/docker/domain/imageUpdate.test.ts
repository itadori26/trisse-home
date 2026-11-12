import { describe, expect, it } from 'vitest';
import { type ImageUpdate, updatesByContainer } from './imageUpdate';

const u = (ref: string, status: ImageUpdate['status'], containers: string[], newerTag = ''): ImageUpdate => ({
  ref,
  containers,
  status,
  newerTag,
  releaseUrl: '',
  checkedAt: null,
  error: '',
});

describe('image updates', () => {
  it('maps updates to containers', () => {
    const m = updatesByContainer([
      u('nginx:1.27', 'update_available', ['web', 'web2']),
      u('postgres:16.1', 'up_to_date', ['db'], '16.4'),
      u('redis:7', 'up_to_date', ['cache']),
    ]);
    expect([...m.keys()].sort()).toEqual(['db', 'web', 'web2']);
    expect(m.get('db')?.newerTag).toBe('16.4');
  });
});
