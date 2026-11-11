import { describe, expect, it } from 'vitest';
import {
  certificateDaysLeft,
  emptyServiceInput,
  isImageIcon,
  statusText,
  validateServiceInput,
  type Service,
} from './service';

const base: Service = {
  ...emptyServiceInput(),
  id: '1',
  name: 'NAS',
  url: 'https://nas',
  position: 0,
  state: 'online',
  httpStatus: 200,
  latencyMs: 12,
  checkedAt: null,
  error: '',
  certExpiresAt: null,
  certExpiresSoon: false,
};

describe('services domain', () => {
  it('validates input', () => {
    expect(validateServiceInput({ ...emptyServiceInput(), url: 'https://x' })).toMatch(/name/);
    expect(validateServiceInput({ ...emptyServiceInput(), name: 'x', url: 'ftp://x' })).toMatch(/URL/);
    expect(validateServiceInput({ ...emptyServiceInput(), name: 'x', url: 'http://x:8080' })).toBeNull();
  });
  it('detects image icons', () => {
    expect(isImageIcon('https://x/icon.png')).toBe(true);
    expect(isImageIcon('🖥️')).toBe(false);
  });
  it('describes the status', () => {
    expect(statusText(base)).toContain('12 ms');
    expect(statusText({ ...base, state: 'offline', error: 'timeout' })).toBe('Offline: timeout');
  });
  it('computes days until certificate expiry', () => {
    const now = new Date('2026-11-11T00:00:00Z');
    expect(certificateDaysLeft({ certExpiresAt: new Date('2026-11-21T12:00:00Z') }, now)).toBe(10);
    expect(certificateDaysLeft({ certExpiresAt: null }, now)).toBeNull();
  });
});
