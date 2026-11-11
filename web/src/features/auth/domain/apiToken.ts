/** A personal API token without its secret. */
export interface ApiToken {
  id: string;
  name: string;
  scopes: string[];
  createdAt: Date;
  expiresAt: Date | null;
  lastUsedAt: Date | null;
  lastUsedIp: string;
}

/** Access level per module in the token form. */
export type Access = 'none' | 'read' | 'write';

/** Modules that can never be reached with a token. */
const excluded = new Set(['auth', 'onboarding']);

/** Modules offered in the scope picker. */
export function scopeModules(modules: Iterable<string>): string[] {
  return [...modules].filter((m) => !excluded.has(m)).sort();
}

/** Turns the per-module choice into scopes such as "todos:write". */
export function toScopes(access: Record<string, Access>): string[] {
  return Object.entries(access)
    .filter(([, a]) => a !== 'none')
    .map(([m, a]) => `${m}:${a}`)
    .sort();
}
