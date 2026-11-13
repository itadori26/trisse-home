/**
 * Copies text to the clipboard. Resolves false when the browser refuses
 * (no permission, insecure context without the Clipboard API).
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}
