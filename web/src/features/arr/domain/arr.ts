import { t } from '../../../shared/i18n';

/** Supported apps (#86). */
export type ArrApp = 'sonarr' | 'radarr' | 'prowlarr';

/** Every app, in the order of the settings. */
export const arrApps: readonly ArrApp[] = ['sonarr', 'radarr', 'prowlarr'];

/** Product name of an app. */
export function appName(a: ArrApp): string {
  return { sonarr: 'Sonarr', radarr: 'Radarr', prowlarr: 'Prowlarr' }[a];
}

/** Default port, for the URL placeholder. */
export function defaultPort(a: ArrApp): number {
  return { sonarr: 8989, radarr: 7878, prowlarr: 9696 }[a];
}

/** A configured instance (the API key is never returned). */
export interface ArrInstance {
  id: string;
  app: ArrApp;
  name: string;
  url: string;
}

/** Editable fields; an empty API key keeps the stored one when updating. */
export interface ArrInstanceInput {
  app: ArrApp;
  name: string;
  url: string;
  apiKey: string;
}

/** An upcoming episode or movie. */
export interface Release {
  instance: string;
  title: string;
  detail: string;
  date: Date;
  hasFile: boolean;
}

/** A download. */
export interface QueueItem {
  instance: string;
  title: string;
  status: string;
  progress: number;
  timeLeftSeconds: number;
  error: string;
}

/** A problem reported by an app. */
export interface HealthIssue {
  instance: string;
  severity: 'notice' | 'warning' | 'error';
  message: string;
}

/** Everything the widget shows. */
export interface ArrOverview {
  upcoming: Release[];
  queue: QueueItem[];
  health: HealthIssue[];
  missing: Record<string, number>;
  errors: string[];
}

/** Sections of the widget. */
export interface ArrWidgetConfig {
  upcoming: boolean;
  queue: boolean;
  health: boolean;
}

/** Reads the widget configuration; every section is shown by default. */
export function parseWidgetConfig(raw: Record<string, unknown>): ArrWidgetConfig {
  return {
    upcoming: raw['upcoming'] !== false,
    queue: raw['queue'] !== false,
    health: raw['health'] !== false,
  };
}

/** Validates the instance form; the key is required only for new instances. */
export function validateInstance(input: ArrInstanceInput, isNew: boolean): string | null {
  if (!/^https?:\/\/\S+$/i.test(input.url.trim())) return t('arr.invalid_url');
  if (isNew && input.apiKey.trim() === '') return t('arr.key_required');
  return null;
}

/** An empty instance form. */
export const emptyInstance = (): ArrInstanceInput => ({ app: 'sonarr', name: '', url: '', apiKey: '' });
