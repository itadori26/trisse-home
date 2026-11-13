import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import ts from 'typescript-eslint';
import svelteConfig from './svelte.config.js';

/**
 * Layer rules: every feature follows domain -> usecase -> adapter -> ui.
 * Inner layers must not import outer layers, and features must only talk to
 * each other through their public index.ts (wired in src/app).
 */
const restrict = (patterns) => ({
  'no-restricted-imports': ['error', { patterns }],
});

export default ts.config(
  {
    ignores: [
      'dist/',
      'node_modules/',
      'test-results/',
      'playwright-report/',
      'src/shared/api-client/generated/',
    ],
  },
  js.configs.recommended,
  ...ts.configs.strict,
  ...svelte.configs.recommended,
  {
    languageOptions: { globals: { ...globals.browser, ...globals.node } },
  },
  {
    files: ['**/*.svelte', '**/*.svelte.ts'],
    languageOptions: {
      parserOptions: {
        projectService: true,
        extraFileExtensions: ['.svelte'],
        parser: ts.parser,
        svelteConfig,
      },
    },
  },
  {
    // Features add their message keys by augmenting an empty interface (#193).
    files: ['src/shared/i18n/registry.ts', 'src/features/*/i18n/en.ts', 'src/app/serviceKit.ts'],
    rules: { '@typescript-eslint/no-empty-object-type': 'off' },
  },
  {
    files: ['src/features/*/domain/**'],
    rules: restrict([
      {
        group: [
          '**/usecase/**',
          '**/adapter/**',
          '**/ui/**',
          'svelte',
          'svelte/*',
          '**/shared/api-client/**',
        ],
        message: 'domain must stay pure',
      },
    ]),
  },
  {
    files: ['src/features/*/usecase/**'],
    rules: restrict([
      {
        group: ['**/adapter/**', '**/ui/**', 'svelte', 'svelte/*', '**/shared/api-client/**'],
        message: 'use cases only know domain and ports',
      },
    ]),
  },
  {
    files: ['src/features/*/adapter/**'],
    rules: restrict([
      { group: ['**/ui/**', 'svelte', 'svelte/*'], message: 'adapters must not depend on the UI' },
    ]),
  },
  {
    files: ['src/features/**'],
    rules: restrict([
      {
        group: ['**/features/*/**', '!./**', '!../**'],
        message: 'features must not import other features; wire them in src/app',
      },
    ]),
  },
);
