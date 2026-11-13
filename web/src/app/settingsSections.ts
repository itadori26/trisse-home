import type { SettingsSection } from './settingsKit';

const found = import.meta.glob<{ section: SettingsSection }>('./settings/*.ts', { eager: true });

/** The sections that features register from src/app/settings, ordered by id (#193). */
export const registeredSections: SettingsSection[] = Object.entries(found)
  .sort(([a], [b]) => a.localeCompare(b))
  .map(([, m]) => m.section);
