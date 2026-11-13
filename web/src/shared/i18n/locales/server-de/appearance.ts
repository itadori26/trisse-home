import type { ServerMessageKey } from '../../../api-client/generated/server-messages/appearance';

/**
 * German translations of the backend error messages of `appearance` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The accent color must look like #3b6ef5.
  srv_1c9c8a99006c73fe: 'Die Akzentfarbe muss wie #3b6ef5 aussehen.',
  // The custom CSS must be at most 64 KiB.
  srv_b410fbcb84d6e168: 'Das eigene CSS darf höchstens 64 KiB groß sein.',
  // The language must be en, de or empty (browser).
  srv_d55a4312396d903f: 'Die Sprache muss en, de oder leer (Browser) sein.',
  // The theme must be system, light or dark.
  srv_d28e94f3f2756a0b: 'Das Design muss system, light oder dark sein.',
};
