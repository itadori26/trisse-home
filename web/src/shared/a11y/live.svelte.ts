import { Announcer } from './announcer';

/** Text of the shared polite live region (#113), rendered by LiveRegion.svelte. */
export const live = $state({ message: '' });

// Clearing first makes screen readers announce a repeated text again.
const announcer = new Announcer((message) => {
  live.message = '';
  setTimeout(() => (live.message = message), 100);
});

/** Politely announces message to screen readers (deduplicated). */
export function announce(message: string): void {
  announcer.announce(message);
}
