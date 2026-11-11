import { describe, expect, it } from 'vitest';
import { scopeModules, toScopes } from './apiToken';

describe('api tokens', () => {
  it('offers modules except auth', () => {
    expect(scopeModules(['todos', 'auth', 'notes', 'onboarding'])).toEqual(['notes', 'todos']);
  });

  it('builds sorted scopes and skips modules without access', () => {
    expect(toScopes({ todos: 'write', notes: 'read', rss: 'none' })).toEqual(['notes:read', 'todos:write']);
  });
});
