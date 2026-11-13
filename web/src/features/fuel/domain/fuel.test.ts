import { describe, expect, it } from 'vitest';
import { formatPrice, isFavourite, parseWidgetConfig, toggleFavourite, validateSettings } from './fuel';

describe('fuel domain', () => {
  it('formats prices like a pump', () => {
    expect(formatPrice(1.739, 'de')).toEqual({ main: '1,73', tenth: '9' });
    expect(formatPrice(1.7, 'en')).toEqual({ main: '1.70', tenth: '0' });
    expect(formatPrice(null, 'de')).toBeNull();
  });
  it('toggles favourites and names them by brand', () => {
    const s = { id: 'a', name: 'Station 12', brand: 'Aral' };
    const added = toggleFavourite([], s);
    expect(added).toEqual([{ id: 'a', name: 'Aral' }]);
    expect(isFavourite(added, 'a')).toBe(true);
    expect(toggleFavourite(added, s)).toEqual([]);
  });
  it('validates the location and radius', () => {
    expect(validateSettings(52.5, 13.4, 5)).toBeNull();
    expect(validateSettings(0, 0, 5)).not.toBeNull();
    expect(validateSettings(52.5, 13.4, 30)).not.toBeNull();
    expect(parseWidgetConfig({ view: 'favourites' }).view).toBe('favourites');
    expect(parseWidgetConfig({}).view).toBe('cheapest');
  });
});
