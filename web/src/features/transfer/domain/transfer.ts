import { t } from '../../../shared/i18n';
/** Identifier of trisse-home export files. */
export const EXPORT_FORMAT = 'trisse-home-export';

/** Result of an import. */
export interface ImportResult {
  imported: string[];
  skipped: string[];
  failed: Record<string, string>;
}

/** Header fields of an export file. */
export interface ExportFile {
  format: string;
  version: number;
  sections: Record<string, unknown>;
}

/** Parses an export file and returns it, or an error message. */
export function parseExportFile(text: string): ExportFile | string {
  let data: unknown;
  try {
    data = JSON.parse(text);
  } catch {
    return t('transfer.invalid_json');
  }
  const f = data as Partial<ExportFile> | null;
  if (
    !f ||
    f.format !== EXPORT_FORMAT ||
    typeof f.version !== 'number' ||
    typeof f.sections !== 'object' ||
    !f.sections
  )
    return t('transfer.not_export');
  return f as ExportFile;
}

/** One-line summary of an import result. */
export function describeResult(r: ImportResult): string {
  const parts = [t('transfer.result_imported', { list: r.imported.join(', ') || t('transfer.nothing') })];
  if (r.skipped.length) parts.push(t('transfer.result_skipped', { list: r.skipped.join(', ') }));
  const failed = Object.entries(r.failed);
  if (failed.length)
    parts.push(t('transfer.result_failed', { list: failed.map(([k, v]) => `${k} (${v})`).join('; ') }));
  return parts.join(' ');
}
