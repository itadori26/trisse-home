import type { ServerMessageKey } from '../../../api-client/generated/server-messages/rss';

/**
 * German translations of the backend error messages of `rss` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The feed URL must start with http:// or https://.
  srv_bbec0c13348226e8: 'Die Feed-URL muss mit http:// oder https:// beginnen.',
  // The feed could not be loaded: {detail}
  srv_4dfea8484179a15f: 'Der Feed konnte nicht geladen werden: {detail}',
  // The feed does not exist.
  srv_5065a34cb1180bb7: 'Der Feed existiert nicht.',
  // The item does not exist.
  srv_fd944f7eadfdb070: 'Der Eintrag existiert nicht.',
  // You are already subscribed to this feed.
  srv_d6a88cffb18a9963: 'Du hast diesen Feed bereits abonniert.',
  // this is not a valid RSS, Atom or JSON feed
  srv_668a30f80af88b85: 'das ist kein gültiger RSS-, Atom- oder JSON-Feed',
  // The OPML file is larger than 2 MB.
  srv_9684d4c2e31b6d2f: 'Die OPML-Datei ist größer als 2 MB.',
  // The file is not an OPML subscription list.
  srv_0d8148b3ba698407: 'Die Datei ist keine OPML-Abo-Liste.',
};
