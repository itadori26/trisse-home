/**
 * Progressive web app helpers. The service worker (public/sw.js) only works
 * in a secure context (HTTPS or localhost); elsewhere these are no-ops.
 */

const SIGNED_IN_KEY = 'trisse-home.signedIn';

/** Registers the service worker in production builds. */
export function registerServiceWorker(): void {
  if (!import.meta.env.PROD || !('serviceWorker' in navigator) || !window.isSecureContext) return;
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').catch(() => {
      // Offline support is optional; the app works without it.
    });
  });
}

/** Remembers whether the user was signed in, so the app can open offline. */
export function rememberSignedIn(signedIn: boolean): void {
  try {
    if (signedIn) localStorage.setItem(SIGNED_IN_KEY, '1');
    else localStorage.removeItem(SIGNED_IN_KEY);
  } catch {
    // storage may be unavailable
  }
}

/** Whether the app may start in offline mode. */
export function canStartOffline(): boolean {
  try {
    return !navigator.onLine && localStorage.getItem(SIGNED_IN_KEY) === '1';
  } catch {
    return false;
  }
}

/** Deletes cached widget data (on logout). */
export async function clearOfflineData(): Promise<void> {
  navigator.serviceWorker?.controller?.postMessage('clear-data');
  if ('caches' in window) await caches.delete('trisse-data-v1').catch(() => false);
}
