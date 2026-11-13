/**
 * German translations of the backend error messages (#102): one file per
 * module next to this one, merged here (#193). A new module adds a file and
 * touches nothing else.
 */
const parts = import.meta.glob<{ serverDe: Record<string, string> }>(['./*.ts', '!./index.ts'], {
  eager: true,
});

export const serverDe: Record<string, string> = Object.assign(
  {},
  ...Object.values(parts).map((m) => m.serverDe),
);
