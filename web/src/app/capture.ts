import type { CaptureForm } from '../features/capture/domain/capture';
import { normaliseUrl, type QuickCommand } from '../features/quickactions/domain/command';
import { extractDue } from '../features/quickactions/domain/dates';

/**
 * The quick command that saves the capture form (#66), or null when
 * something is missing. Wired here because the capture page creates its
 * items through the quick actions, which the capture feature does not know.
 * To-dos take a due date from the title ("… tomorrow"); notes keep the
 * shared text and link below the title.
 */
export function captureCommand(f: CaptureForm, today: Date): QuickCommand | null {
  const title = f.title.trim();
  switch (f.target) {
    case 'bookmark': {
      const url = normaliseUrl(f.url);
      if (!url) return null;
      return { kind: 'bookmark', url, title: title || new URL(url).hostname };
    }
    case 'todo': {
      if (!title) return null;
      const { text, due } = extractDue(title, today);
      return { kind: 'todo', title: text || title, due };
    }
    case 'note': {
      const body = [title, f.text.trim(), f.url.trim()].filter(Boolean).join('\n');
      return body ? { kind: 'note', title: body } : null;
    }
  }
}
