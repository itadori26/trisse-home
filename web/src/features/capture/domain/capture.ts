/** What arrives at /capture from the bookmarklet, the share target or the browser extension (#66). */
export interface Shared {
  url: string;
  title: string;
  text: string;
}

export type CaptureTarget = 'bookmark' | 'todo' | 'note';
export const CAPTURE_TARGETS: readonly CaptureTarget[] = ['bookmark', 'todo', 'note'];

const MAX = { url: 2000, title: 200, text: 4000 };
const urlPattern = /https?:\/\/[^\s<>"']+/i;

/**
 * Reads the query of /capture. Android puts a shared link into "text" more
 * often than into "url", so the first http(s) link of the text is used when
 * there is no url, and removed from the text.
 */
export function parseShared(search: string): Shared {
  const q = new URLSearchParams(search);
  const get = (k: keyof Shared) => (q.get(k) ?? '').trim().slice(0, MAX[k]);
  let url = get('url');
  let text = get('text');
  if (!url) {
    const m = urlPattern.exec(text);
    if (m) {
      url = m[0];
      text = text
        .replace(m[0], '')
        .replace(/\s{2,}/g, ' ')
        .trim();
    }
  }
  if (!/^https?:\/\//i.test(url)) url = '';
  return { url, title: get('title'), text };
}

/** A link becomes a bookmark, anything else a note. */
export const defaultTarget = (s: Shared): CaptureTarget => (s.url ? 'bookmark' : 'note');

/** The form fields of the capture page. */
export interface CaptureForm {
  target: CaptureTarget;
  title: string;
  url: string;
  text: string;
}

/** The prefilled form for what was shared. */
export function initialForm(s: Shared): CaptureForm {
  const target = defaultTarget(s);
  return {
    target,
    url: s.url,
    title: s.title || (target === 'note' ? '' : s.text),
    text: target === 'note' ? s.text : '',
  };
}

/** The bookmarklet that opens /capture of this server with the current page. */
export function bookmarklet(origin: string): string {
  return (
    `javascript:(()=>{const e=encodeURIComponent;window.open('${origin}/capture?url='+e(location.href)+'&title='+e(document.title)` +
    `+'&text='+e(String(getSelection())),'trisse-capture','width=520,height=640')})()`
  );
}
