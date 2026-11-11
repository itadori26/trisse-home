import { describe, expect, it } from 'vitest';
import { describeResult, parseExportFile } from './transfer';

describe('transfer domain', () => {
  it('validates export files', () => {
    expect(parseExportFile('nope')).toMatch(/JSON/);
    expect(parseExportFile('{"format":"other"}')).toMatch(/not a trisse-home/);
    expect(parseExportFile('{"format":"trisse-home-export","version":1,"sections":{}}')).toMatchObject({
      version: 1,
    });
  });
  it('summarises results', () => {
    expect(describeResult({ imported: ['notes'], skipped: ['calendar'], failed: { todos: 'broken' } })).toBe(
      'Imported: notes. Skipped (module disabled or unknown): calendar. Failed: todos (broken).',
    );
  });
});
