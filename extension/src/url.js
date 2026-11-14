// Pure helpers of the extension (#65), tested with node --test.

/**
 * Normalises the dashboard address entered in the options: http(s) only,
 * no credentials, query or fragment; the trailing slash is dropped. Returns
 * null for anything else.
 */
export function normaliseBase(raw) {
  let u;
  try {
    u = new URL(String(raw ?? '').trim());
  } catch {
    return null;
  }
  if ((u.protocol !== 'http:' && u.protocol !== 'https:') || u.username || u.password || u.search || u.hash) return null;
  return (u.origin + u.pathname).replace(/\/+$/, '');
}

/** The dashboard with the search field focused (handled by the frontend). */
export const dashboardUrl = (base) => `${base}/?focus=search`;

/** The health check used to tell "unreachable" from "slow". */
export const healthUrl = (base) => `${base}/healthz`;

/**
 * The quick capture page (#66) prefilled with a tab; only http(s) pages are
 * handed over, so browser-internal pages do not leak their URLs.
 */
export function captureUrl(base, tab) {
  const q = new URLSearchParams();
  const url = String(tab?.url ?? '');
  if (/^https?:\/\//i.test(url)) q.set('url', url);
  if (tab?.title) q.set('title', String(tab.title).slice(0, 200));
  if (tab?.selection) q.set('text', String(tab.selection).slice(0, 4000));
  return `${base}/capture?${q}`;
}
