import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * Services are part of the start-up code. If one imports a feature index that
 * also exports a Svelte component, that component (usually the settings UI)
 * lands in the initial bundle (#230). Services import adapter and use cases
 * directly instead.
 */
describe('service registrations (#230)', () => {
  it('never import a feature module that exports Svelte components', () => {
    const dir = join(__dirname, 'services');
    const offenders = readdirSync(dir)
      .filter((f) => f.endsWith('.ts') && !f.endsWith('.test.ts'))
      .flatMap((file) => {
        const source = readFileSync(join(dir, file), 'utf8');
        return [...source.matchAll(/from '\.\.\/\.\.\/(features\/[^']+)'/g)]
          .map((m) => m[1] ?? '')
          .filter((target) => {
            const candidates = [`${target}.ts`, `${target}/index.ts`];
            return candidates.some((c) => {
              try {
                return /\.svelte['"]/.test(readFileSync(join(__dirname, '..', c), 'utf8'));
              } catch {
                return false;
              }
            });
          })
          .map((target) => `${file} imports ${target}`);
      });
    expect(offenders).toEqual([]);
  });
});
