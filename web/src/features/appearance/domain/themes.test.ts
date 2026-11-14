import { describe, expect, it, vi } from 'vitest';
import { previewStyle } from './themes';

describe('theme preview (#111)', () => {
  it('only lets colors and radii through', () => {
    expect(
      previewStyle({
        tokens: {
          '--color-bg': '#2e3440',
          '--radius': '6px',
          '--color-text': 'red; background: url(https://evil.example)',
          color: '#fff',
        },
      }),
    ).toBe('--color-bg: #2e3440; --radius: 6px');
  });
});

describe('theme import (#111)', () => {
  it('refuses files above 16 KB before uploading', async () => {
    const { AppearanceUseCases } = await import('../usecase/appearanceUseCases');
    const importTheme = vi.fn().mockResolvedValue({ id: 'x' });
    const uc = new AppearanceUseCases({ importTheme } as unknown as ConstructorParameters<
      typeof AppearanceUseCases
    >[0]);
    await expect(uc.importTheme({ size: (16 << 10) + 1, text: () => Promise.resolve('') })).rejects.toThrow();
    await uc.importTheme({ size: 2, text: () => Promise.resolve('{}') });
    expect(importTheme).toHaveBeenCalledWith('{}');
  });
});
