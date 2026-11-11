/* trisse-home service worker: app shell and last-known widget data offline.
 *
 * Security rules (see README "Progressive Web App"):
 * - only GET requests of the same origin are ever cached;
 * - authentication, settings, exports, backups, logs and anything else not on
 *   the allow-list below is never cached (they may carry tokens or secrets);
 * - the page asks the worker to delete the data cache on logout.
 */
const SHELL = 'trisse-shell-v1';
const DATA = 'trisse-data-v1';

/** Read-only widget data that may be shown offline. */
const DATA_ALLOW = [
  /^\/api\/modules$/,
  /^\/api\/appearance$/,
  /^\/api\/dashboard$/,
  /^\/api\/weather(\/warnings)?$/,
  /^\/api\/bookmarks$/,
  /^\/api\/notes$/,
  /^\/api\/todos$/,
  /^\/api\/services(\/history)?$/,
  /^\/api\/system(\/history)?$/,
  /^\/api\/rss\/items$/,
  /^\/api\/calendar\/events$/,
  /^\/api\/docker\/containers$/,
  /^\/api\/adguard\/stats$/,
  /^\/api\/waste\/upcoming$/,
  /^\/api\/transit\/departures$/,
  /^\/api\/nextcloud\/status$/,
  /^\/api\/fritzbox\/status$/,
  /^\/api\/wireguard\/peers$/,
  /^\/api\/parcels$/,
  /^\/api\/backup-status\/sources$/,
  /^\/api\/shopping\/items$/,
];

self.addEventListener('install', (event) => {
  event.waitUntil(caches.open(SHELL).then((c) => c.addAll(['/', '/favicon.svg', '/manifest.webmanifest'])));
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) => Promise.all(keys.filter((k) => k !== SHELL && k !== DATA).map((k) => caches.delete(k))))
      .then(() => self.clients.claim()),
  );
});

self.addEventListener('message', (event) => {
  if (event.data === 'clear-data') event.waitUntil(caches.delete(DATA));
});

/** Network first; on failure the cached copy (if any). */
async function networkFirst(request, cacheName, cacheKey = request) {
  const cache = await caches.open(cacheName);
  try {
    const response = await fetch(request);
    if (response.ok && !response.headers.get('Set-Cookie')) await cache.put(cacheKey, response.clone());
    return response;
  } catch (err) {
    const cached = await cache.match(cacheKey);
    if (cached) return cached;
    throw err;
  }
}

/** Cache first for immutable, hashed build assets. */
async function cacheFirst(request) {
  const cache = await caches.open(SHELL);
  const cached = await cache.match(request);
  if (cached) return cached;
  const response = await fetch(request);
  if (response.ok) await cache.put(request, response.clone());
  return response;
}

self.addEventListener('fetch', (event) => {
  const { request } = event;
  if (request.method !== 'GET') return;
  const url = new URL(request.url);
  if (url.origin !== self.location.origin) return;
  if (request.mode === 'navigate') {
    event.respondWith(networkFirst(request, SHELL, '/'));
  } else if (url.pathname.startsWith('/assets/') || url.pathname.startsWith('/icons/')) {
    event.respondWith(cacheFirst(request));
  } else if (DATA_ALLOW.some((re) => re.test(url.pathname))) {
    event.respondWith(networkFirst(request, DATA));
  }
  // Everything else (auth, settings, exports, …) goes to the network untouched.
});
