/** Result of the image update check (#107). */
export type UpdateStatus = 'unknown' | 'up_to_date' | 'update_available' | 'pinned' | 'local' | 'error';

/** One running image with its update state. */
export interface ImageUpdate {
  ref: string;
  containers: string[];
  status: UpdateStatus;
  newerTag: string;
  releaseUrl: string;
  checkedAt: Date | null;
  error: string;
}

/** A registry login; the password is never sent back. */
export interface RegistryLogin {
  host: string;
  username: string;
  hasPassword: boolean;
}

/** Settings of the update check (off by default). */
export interface UpdateSettings {
  enabled: boolean;
  semverTags: boolean;
  registries: RegistryLogin[];
}

/** A registry login being edited; an empty password keeps the stored one. */
export interface RegistryInput {
  host: string;
  username: string;
  password: string;
}

/** Images with an update (newer digest or newer tag) by container name. */
export function updatesByContainer(list: readonly ImageUpdate[]): Map<string, ImageUpdate> {
  const out = new Map<string, ImageUpdate>();
  for (const u of list) {
    if (!hasUpdate(u)) continue;
    for (const c of u.containers) out.set(c, u);
  }
  return out;
}

/** Whether the image has a newer digest for its tag or a newer tag. */
export function hasUpdate(u: ImageUpdate): boolean {
  return u.status === 'update_available' || u.newerTag !== '';
}
