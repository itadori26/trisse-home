import { describe, expect, it } from 'vitest';
import {
  emptyInput,
  formatValue,
  intervalLabel,
  parseWidgetConfig,
  toInput,
  validateInput,
} from './apiwidget';

describe('JSON widget domain (#48)', () => {
  const input = () => ({
    ...emptyInput(),
    name: 'NAS',
    url: 'http://nas.lan/api',
    extractions: [{ label: 'Temp', path: '$.t', unit: '°C', warn: null, crit: null, below: false }],
  });

  it('validates the form', () => {
    expect(validateInput(input(), false)).toBeNull();
    expect(validateInput({ ...input(), name: ' ' }, false)).not.toBeNull();
    expect(validateInput({ ...input(), url: 'nas.lan' }, false)).not.toBeNull();
    expect(validateInput({ ...input(), extractions: [] }, false)).not.toBeNull();
    expect(
      validateInput({ ...input(), extractions: input().extractions.map((x) => ({ ...x, path: '' })) }, false),
    ).not.toBeNull();
    // A header needs a value, unless one is stored already.
    expect(validateInput({ ...input(), headerName: 'X-Api-Key' }, false)).not.toBeNull();
    expect(validateInput({ ...input(), headerName: 'X-Api-Key' }, true)).toBeNull();
  });

  it('formats numbers in the locale and keeps texts', () => {
    expect(formatValue({ text: '21.456', number: 21.456, unit: '°C' }, 'de')).toBe('21,46 °C');
    expect(formatValue({ text: '1234567', number: 1234567, unit: '' }, 'en')).toBe('1,234,567');
    expect(formatValue({ text: 'healthy', number: null, unit: '' }, 'en')).toBe('healthy');
  });

  it('reads the widget config and edits a stored endpoint without its secret', () => {
    expect(parseWidgetConfig({ endpointId: 'a' })).toEqual({ endpointId: 'a' });
    expect(parseWidgetConfig({ endpointId: 3 })).toEqual({ endpointId: '' });
    const form = toInput({
      id: 'a',
      name: 'NAS',
      url: 'http://nas/api',
      headerName: 'X-Api-Key',
      hasHeader: true,
      intervalSeconds: 60,
      allowPublic: false,
      extractions: input().extractions,
    });
    expect(form.headerValue).toBe('');
    expect(form.headerName).toBe('X-Api-Key');
    expect(intervalLabel(30)).toBe('30 s');
    expect(intervalLabel(300)).toBe('5 min');
    expect(intervalLabel(21_600)).toBe('6 h');
  });
});
