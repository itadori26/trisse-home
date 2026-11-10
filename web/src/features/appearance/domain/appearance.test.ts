import { describe, expect, it } from 'vitest';
import { cssVariables, defaultAppearance } from './appearance';

describe('cssVariables', () => {
  it('is empty for defaults', () => {
    expect(cssVariables(defaultAppearance)).toEqual({});
  });
  it('maps accent and background', () => {
    const vars = cssVariables({
      ...defaultAppearance,
      accentColor: '#112233',
      backgroundImage: 'https://x.y/bg.jpg',
    });
    expect(vars['--color-accent']).toBe('#112233');
    expect(vars['--background-image']).toBe('url("https://x.y/bg.jpg")');
  });
  it('ignores values that could break out of url()', () => {
    expect(
      cssVariables({ ...defaultAppearance, backgroundImage: 'https://x.y/a.png") ; color: red' }),
    ).toEqual({});
  });
});
