import DOMPurify from 'dompurify';
import { marked } from 'marked';

/**
 * Renders Markdown to sanitised HTML. Raw HTML in the source is removed by
 * DOMPurify, and links open in a new tab without leaking the referrer.
 */
export function renderMarkdown(source: string): string {
  const html = marked.parse(source, { async: false, gfm: true, breaks: true });
  const clean = DOMPurify.sanitize(html, { USE_PROFILES: { html: true } });
  const doc = new DOMParser().parseFromString(clean, 'text/html');
  for (const a of doc.querySelectorAll('a')) {
    a.setAttribute('target', '_blank');
    a.setAttribute('rel', 'noopener noreferrer');
  }
  return doc.body.innerHTML;
}
