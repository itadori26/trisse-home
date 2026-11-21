/**
 * The plugin API of the web app (ADR 0021). The frontend of ee/ is a separate
 * package that the app loads at runtime; this file is the whole contract
 * between the two. Keep it small and stable: a plugin built against version
 * 1 must keep working while the core changes.
 */

/**
 * Version of this contract. Version 2 added compact views of plugin widgets
 * and translations of the plugin's backend messages (#632); version 3 added
 * steps of the first-start flow (#671).
 */
export const PLUGIN_API_VERSION = 3;

/**
 * Oldest version the loader still accepts: later versions only add optional
 * fields, so a plugin built against it keeps working.
 */
export const MIN_PLUGIN_API_VERSION = 1;

/** Access to a licensed feature, as GET /api/licence/features reports it. */
export type FeatureLevel = 'full' | 'read' | 'none';

/** The API client a plugin uses; session and CSRF are handled by the core. */
export interface PluginHttp {
  get<T>(path: string): Promise<T>;
  post<T>(path: string, body?: unknown): Promise<T>;
  put<T>(path: string, body?: unknown): Promise<T>;
  delete<T = void>(path: string): Promise<T>;
  /** PUTs a file as the raw body (images and the like). */
  putBlob<T>(path: string, blob: Blob): Promise<T>;
}

/** What a plugin screen gets from the core besides its own props. */
export interface PluginContext {
  http: PluginHttp;
  /** Active UI language, e.g. "en" or "de". */
  language: string;
  /** The signed-in account. */
  account: { owner: boolean; admin: boolean; member: boolean; guest: boolean };
  /** Access to a licence feature. */
  level(feature: string): FeatureLevel;
  /** Turns a server error into a message for the user, translated. */
  errorMessage(err: unknown): string;
}

/**
 * Something a plugin draws into a host element the core provides. It returns
 * a function that removes it again. Plugins bundle their own Svelte runtime,
 * so they mount into the element instead of joining the core's component tree.
 */
export interface Mountable {
  mount(host: HTMLElement, props: Record<string, unknown>, context: PluginContext): () => void;
}

/** A text in every language the plugin supports; English is the fallback. */
export type Texts = { en: string } & Record<string, string>;

/**
 * Where a plugin widget's compact view reads its short state:
 * GET /api/<module>/compact with params as the query (#595, #614).
 */
export interface PluginCompactSource {
  module: string;
  params?: Record<string, string>;
}

/** A widget kind contributed by a plugin. */
export interface PluginWidget {
  kind: string;
  /** Licence feature that unlocks it. */
  feature: string;
  /** Picker category; organisation features usually belong to "everyday". */
  category: 'everyday' | 'information';
  title: Texts;
  description: Texts;
  keywords?: string[];
  defaultSize: { w: number; h: number };
  /** One of the core's icon names; unknown names fall back to a generic icon. */
  icon: string;
  /** Receives { config }. */
  view: Mountable;
  /** Receives { config, onsave(config) }. */
  settings?: Mountable;
  /**
   * The compact state source for a configuration, shown in small and strip
   * tiles; undefined keeps the full view at every size. Since version 2.
   */
  compact?: (config: Record<string, unknown>) => PluginCompactSource | undefined;
  /** Example of the compact view for the size previews of the picker (#596). Since version 2. */
  compactExample?: Texts;
}

/** A section of the settings dialog contributed by a plugin. */
export interface PluginSection {
  id: string;
  /** Licence feature it belongs to; omitted for sections that are always open. */
  feature?: string;
  label: Texts;
  /** Shown instead of the form while the feature is off. */
  description: Texts;
  /** Only admins see it. */
  adminOnly?: boolean;
  /**
   * Task area of the settings start page (#600): "account", "appearance",
   * "notifications", "dashboard", "connections", "people", "data" or
   * "system". Optional; unknown values and older plugins land in "people".
   */
  area?: string;
  view: Mountable;
}

/**
 * A command the start page search offers when the typed text matches one of
 * its keywords (#453), e.g. "office" → "Set my status: Office". The core
 * does the matching and hides the action from guests and while its feature
 * is not fully licensed, so a plugin only describes it.
 */
export interface PluginSearchAction {
  /** Unique within the plugin. */
  id: string;
  /** Licence feature it needs at level "full". */
  feature?: string;
  /**
   * Words that offer the action, in every supported language. The typed
   * text matches a keyword it equals or starts (from two letters on),
   * ignoring case and surrounding spaces.
   */
  keywords: string[];
  title: Texts;
  /** Shown muted next to the title, e.g. the feature's name. */
  detail?: Texts;
  /** Shown after the action ran; the title is used without it. */
  done?: Texts;
  /** Runs the action; a rejection is shown as an error message. */
  run(context: PluginContext): Promise<void>;
}

/**
 * A step of the first-start flow (#671), shown after the core's own steps
 * and before the template choice. The core draws the heading, the progress
 * and the Back, Skip and Next buttons; the view receives
 * { setSubmit(save: () => Promise<void>) } and registers what Next saves.
 * A rejected save keeps the step open and shows the error.
 */
export interface PluginOnboardingStep {
  /** Unique within the plugin. */
  id: string;
  /** Licence feature it needs at level "full"; the step is left out otherwise. */
  feature?: string;
  /** Only the owner sees it; invited members get the personal steps only. */
  ownerOnly?: boolean;
  title: Texts;
  view: Mountable;
}

/**
 * What a plugin module exports as `plugin`. Fields added after version 1
 * shipped are optional, so plugins built against the first version still
 * load and an older core ignores what it does not know.
 */
export interface WaylightPlugin {
  apiVersion: number;
  name: string;
  widgets: PluginWidget[];
  settings: PluginSection[];
  /** Commands for the search field (#453); optional. */
  searchActions?: PluginSearchAction[];
  /** Steps of the first-start flow (#671); optional. Since version 3. */
  onboardingSteps?: PluginOnboardingStep[];
  /**
   * Translations of the plugin's backend messages (errors and compact
   * states) by language and message key, e.g. { de: { srv_…: '…' } }.
   * English needs none: the server sends it. Since version 2.
   */
  serverMessages?: Record<string, Record<string, string>>;
}

/** Picks the text for a language, falling back to English. */
export const pick = (texts: Texts, language: string): string => texts[language] ?? texts.en;
