import { describe, expect, it, vi } from 'vitest';
import { expiresIn, formatSize, validateFile } from './drop';
import { DropUseCases, type DropGateway } from '../usecase/dropUseCases';

describe('drop zone (#101)', () => {
  it('describes the expiry and sizes', () => {
    const now = new Date(2026, 10, 14, 8, 0);
    const at = (h: number) => ({ expiresAt: new Date(now.getTime() + h * 3600_000) });
    expect(expiresIn(at(0.5), now)).toBe('in 30 min');
    expect(expiresIn(at(23), now)).toBe('in 23 h');
    expect(expiresIn(at(168), now)).toBe('in 7 days');
    expect(formatSize(3 * 1024 * 1024, 'en')).toBe('3 MB');
    expect(validateFile({ size: 0 })).not.toBeNull();
    expect(validateFile({ size: (25 << 20) + 1 })).not.toBeNull();
    expect(validateFile({ size: 10 })).toBeNull();
  });

  it('uploads files in order and stops at a refused one', async () => {
    const g = {
      list: vi.fn(),
      addText: vi.fn(),
      addFile: vi.fn().mockResolvedValue({}),
      remove: vi.fn(),
      fileUrl: vi.fn(),
    } satisfies DropGateway;
    const uc = new DropUseCases(g);
    const file = (name: string, size: number) => ({ name, size }) as File;
    await uc.addFiles([file('a', 1), file('b', 2)], 24);
    expect(g.addFile).toHaveBeenCalledTimes(2);
    await expect(uc.addFiles([file('c', 0)], 24)).rejects.toThrow('c:');
  });
});
