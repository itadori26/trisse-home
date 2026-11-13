import { readdirSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { gzipSync } from 'node:zlib';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import type { Plugin } from 'vite';
import { defineConfig } from 'vitest/config';

/** Text formats that compress well; images and fonts are already compressed. */
const compressible = /\.(?:js|css|html|svg|json|webmanifest|txt)$/;

/**
 * Replaces every compressible file of the build output by name.gz. The Go
 * binary embeds the output, so it carries about a third of the size; the SPA
 * handler sends the .gz as is to clients that accept gzip and decompresses it
 * for the others (internal/platform/spa).
 */
function gzipOutput(): Plugin {
  let outDir = '';
  const walk = (dir: string): void => {
    for (const entry of readdirSync(dir)) {
      const file = join(dir, entry);
      if (statSync(file).isDirectory()) walk(file);
      else if (compressible.test(entry)) {
        writeFileSync(`${file}.gz`, gzipSync(readFileSync(file), { level: 9 }));
        rmSync(file);
      }
    }
  };
  return {
    name: 'gzip-output',
    apply: 'build',
    configResolved(config) {
      outDir = resolve(config.root, config.build.outDir);
    },
    closeBundle() {
      walk(outDir);
    },
  };
}

/** Backend address used by the dev server proxy (`npm run dev`). */
const backend = process.env.BACKEND_URL ?? 'http://127.0.0.1:8090';

export default defineConfig({
  plugins: [svelte(), gzipOutput()],
  build: {
    // The Go binary embeds the build output via internal/platform/webui.
    outDir: '../internal/platform/webui/dist',
    emptyOutDir: true,
    sourcemap: false,
  },
  server: {
    proxy: { '/api': backend, '/healthz': backend },
  },
  test: {
    include: ['src/**/*.test.ts'],
    environment: 'node',
    // Catalogues load lazily (#116); tests start like the app, after loading.
    setupFiles: ['src/test-setup.ts'],
  },
});
