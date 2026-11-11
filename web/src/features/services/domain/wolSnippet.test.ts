import { describe, expect, it } from 'vitest';
import { validBroadcast, wolEnvSnippet } from './wolSnippet';

describe('wolEnvSnippet', () => {
  it('enables the relay profile', () => {
    expect(wolEnvSnippet({ setup: 'relay', broadcast: '', dockerActions: false })).toBe(
      'COMPOSE_PROFILES=wol',
    );
    expect(wolEnvSnippet({ setup: 'relay', broadcast: '192.168.178.255:9', dockerActions: true })).toBe(
      'COMPOSE_PROFILES=wol,docker-actions\nWOL_BROADCAST=192.168.178.255:9',
    );
  });

  it('disables the relay in the host network', () => {
    expect(wolEnvSnippet({ setup: 'host-network', broadcast: '', dockerActions: false })).toBe('WOL_RELAY=');
  });

  it('validates the broadcast address', () => {
    expect(validBroadcast('10.0.0.255:9')).toBe(true);
    expect(validBroadcast('10.0.0.256:9')).toBe(false);
    expect(validBroadcast('10.0.0.255')).toBe(false);
    expect(wolEnvSnippet({ setup: 'relay', broadcast: 'x', dockerActions: false })).toBeNull();
  });
});
