import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vitest/config';

/** Backend address used by the dev server proxy (`npm run dev`). */
const backend = process.env.BACKEND_URL ?? 'http://127.0.0.1:8090';

export default defineConfig({
  plugins: [svelte()],
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
  },
});
