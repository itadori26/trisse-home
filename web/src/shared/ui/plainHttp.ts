/**
 * Whether one of the entered addresses uses plain http (#191): integrations
 * accept it for LAN devices, but passwords and tokens then travel
 * unencrypted, so the settings show a warning. Several addresses may be
 * given one per line (CalDAV calendars).
 */
export function usesPlainHttp(urls: string): boolean {
  return urls.split('\n').some((u) => /^\s*http:\/\//i.test(u));
}
