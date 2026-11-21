import { describe, expect, it, vi } from 'vitest';
import { GiftsUseCases, type GiftsGateway } from './giftsUseCases';

describe('GiftsUseCases', () => {
  it('cleans the input and toggles bought with the full idea', async () => {
    const g: GiftsGateway = { list: vi.fn(), create: vi.fn(), update: vi.fn(), remove: vi.fn() };
    const uc = new GiftsUseCases(g);
    await uc.save(null, { person: ' Grandma  Hilde ', title: ' Book ', link: '', note: ' ', bought: false });
    expect(g.create).toHaveBeenCalledWith({
      person: 'Grandma Hilde',
      title: 'Book',
      link: '',
      note: '',
      bought: false,
    });
    await uc.setBought({ id: '1', person: 'Ben', title: 'Ball', link: '', note: 'red', bought: false }, true);
    expect(g.update).toHaveBeenCalledWith('1', {
      person: 'Ben',
      title: 'Ball',
      link: '',
      note: 'red',
      bought: true,
    });
  });
});
