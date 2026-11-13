/** Connection settings as shown in the UI (the API key is never returned). */
export interface ImmichSettings {
  configured: boolean;
  url: string;
  hasApiKey: boolean;
}

/** "On this day" some years ago; thumbnails are proxied by the backend. */
export interface Memory {
  yearsAgo: number;
  assetIds: string[];
}

/** What the widget shows. */
export interface ImmichOverview {
  version: string;
  photos: number;
  videos: number;
  diskUsed: number;
  diskTotal: number;
  diskPercent: number;
  memories: Memory[];
}

/** URL of the proxied thumbnail of a memory photo. */
export function thumbnailUrl(assetId: string): string {
  return `/api/immich/thumbnails/${encodeURIComponent(assetId)}`;
}

/** Widget configuration. */
export interface ImmichWidgetConfig {
  showMemories: boolean;
}

/** Reads the widget configuration; memories are shown by default. */
export function parseWidgetConfig(raw: Record<string, unknown>): ImmichWidgetConfig {
  return { showMemories: raw['showMemories'] !== false };
}
