import type { ServerMessageKey } from '../../../api-client/generated/server-messages/gamedeals';

/**
 * German translations of the backend error messages of `gamedeals` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Could not load the Steam wishlist (is it public?): {detail}
  srv_1cefdc3935bd67ea: 'Die Steam-Wunschliste konnte nicht geladen werden (ist sie öffentlich?): {detail}',
  // The Steam ID must be the 17-digit number from your profile URL.
  srv_78609f4b2494d390: 'Die Steam-ID muss die 17-stellige Zahl aus deiner Profil-URL sein.',
  // The country must be a two-letter code such as DE.
  srv_54b6e0f5f98286bf: 'Das Land muss ein zweistelliger Code wie DE sein.',
  // The minimum discount must be between 0 and 100 %.
  srv_f32556d4b742441a: 'Der Mindestrabatt muss zwischen 0 und 100 % liegen.',
};
