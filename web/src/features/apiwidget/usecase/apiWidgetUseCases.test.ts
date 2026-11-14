import { describe, expect, it, vi } from 'vitest';
import { emptyInput, type Endpoint } from '../domain/apiwidget';
import { ApiWidgetUseCases, type ApiWidgetGateway } from './apiWidgetUseCases';

const gw = () =>
  ({
    endpoints: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
    preview: vi.fn(),
    values: vi.fn(),
  }) satisfies ApiWidgetGateway;

const form = () => ({
  ...emptyInput(),
  name: ' NAS ',
  url: ' http://nas/api ',
  headerName: ' X-Api-Key ',
  extractions: [{ label: ' Temp ', path: ' $.t ', unit: '', warn: Number.NaN, crit: 5, below: true }],
});

describe('ApiWidgetUseCases', () => {
  it('cleans the form and creates or updates', async () => {
    const g = gw();
    const uc = new ApiWidgetUseCases(g);
    await expect(uc.save(form())).rejects.toThrow(); // header without value
    const stored = { id: 'a', hasHeader: true } as Endpoint;
    await uc.save(form(), stored);
    expect(g.update).toHaveBeenCalledWith('a', {
      ...form(),
      name: 'NAS',
      url: 'http://nas/api',
      headerName: 'X-Api-Key',
      extractions: [{ label: 'Temp', path: '$.t', unit: '', warn: null, crit: 5, below: true }],
    });
    await uc.save({ ...form(), headerName: '' });
    expect(g.create).toHaveBeenCalledTimes(1);
  });

  it('previews with the stored endpoint', async () => {
    const g = gw();
    await new ApiWidgetUseCases(g).preview(form(), { id: 'a', hasHeader: true } as Endpoint);
    expect(g.preview.mock.calls[0]?.[1]).toBe('a');
  });
});
