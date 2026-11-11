import { describe, expect, it } from 'vitest';
import { eventLabel, isWarning } from './securityEvent';

describe('security events', () => {
  it('maps types to message keys with a fallback', () => {
    expect(eventLabel('login_locked')).toBe('auth.event.login_locked');
    expect(eventLabel('something_new')).toBe('auth.event.other');
  });

  it('highlights suspicious events', () => {
    expect(isWarning('new_device')).toBe(true);
    expect(isWarning('login_succeeded')).toBe(false);
  });
});
