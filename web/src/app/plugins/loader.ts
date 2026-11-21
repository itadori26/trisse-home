import { errorMessage, type HttpClient } from '../../shared/api-client/http';
import { locale, registerServerMessages } from '../../shared/i18n';
import { compactSource } from '../../features/compact/domain/compact';
import type { IconName } from '../../shared/ui/icons/paths';
import { iconPaths } from '../../shared/ui/icons/paths';
import type { AuthUseCases } from '../../features/auth';
import type { SearchAction } from '../searchActions';
import type { SettingsSection } from '../settingsKit';
import { isSettingsArea, type SettingsArea } from '../settingsNav';
import type { WidgetDefinition } from '../widgetKit';
import {
  MIN_PLUGIN_API_VERSION,
  PLUGIN_API_VERSION,
  pick,
  type FeatureLevel,
  type Mountable,
  type PluginContext,
  type PluginOnboardingStep,
  type PluginSearchAction,
  type PluginSection,
  type PluginWidget,
  type WaylightPlugin,
} from './api';
import LockedFeature from './LockedFeature.svelte';
import PluginHost from './PluginHost.svelte';

/** Where the loaded plugins put their contributions. */
export interface PluginTargets {
  widgets: WidgetDefinition[];
  sections: SettingsSection[];
  /** Commands for the search field (#453); left out where nobody searches. */
  searchActions?: SearchAction[];
  /** Steps of the first-start flow (#671); left out where there is none. */
  onboardingSteps?: OnboardingStep[];
}

/** A plugin step of the first-start flow, ready for the core to host. */
export interface OnboardingStep {
  id: string;
  title: () => string;
  mountable: Mountable;
  context: PluginContext;
}

/** How the loader reaches a plugin module; tests replace it. */
export type ImportPlugin = (url: string) => Promise<{ plugin?: WaylightPlugin }>;

const browserImport: ImportPlugin = (url) => import(/* @vite-ignore */ url);

/** GETs a JSON endpoint that only exists with ee/; a missing one means "nothing". */
async function optional<T>(http: HttpClient, path: string, fallback: T): Promise<T> {
  try {
    return await http.get<T>(path);
  } catch {
    return fallback;
  }
}

const knownIcon = (name: string): IconName => (name in iconPaths ? (name as IconName) : 'box');

/** Wraps a plugin screen as a lazily "loaded" PluginHost with its props. */
function hosted(m: Mountable, context: PluginContext) {
  return {
    load: () => Promise.resolve(PluginHost as never),
    props: (config: Record<string, unknown>) => ({ mountable: m, context, props: { config } }),
  };
}

function toWidget(w: PluginWidget, context: PluginContext): WidgetDefinition {
  return {
    type: {
      kind: w.kind,
      category: w.category,
      keywords: w.keywords ?? [],
      get title() {
        return pick(w.title, locale());
      },
      get description() {
        return pick(w.description, locale());
      },
      defaultSize: w.defaultSize,
    },
    icon: knownIcon(w.icon),
    view: hosted(w.view, context),
    settings: w.settings ? hosted(w.settings, context) : undefined,
    compact: w.compact
      ? (config) => {
          const source = w.compact?.(config);
          return source ? compactSource(source.module, source.params) : undefined;
        }
      : undefined,
    compactExample: w.compactExample ? () => pick(w.compactExample ?? { en: '' }, locale()) : undefined,
  };
}

/** Whether the loader understands a plugin built against version v. */
export const supportedVersion = (v: unknown): boolean =>
  typeof v === 'number' && v >= MIN_PLUGIN_API_VERSION && v <= PLUGIN_API_VERSION;

/**
 * Where a plugin section goes when it names no known area: plugin sections
 * are mostly about running an organisation, i.e. people and their access.
 */
const PLUGIN_AREA: SettingsArea = 'people';

/**
 * A plugin section is listed even while its feature is off (ADR 0021): it
 * then shows what the feature does and that it needs a licence key.
 */
function toSection(s: PluginSection, context: PluginContext): SettingsSection {
  const locked = () => !!s.feature && context.level(s.feature) === 'none';
  return {
    id: `plugin-${s.id}`,
    label: 'settings.plugin_section',
    title: () => pick(s.label, locale()),
    area: isSettingsArea(s.area) ? s.area : PLUGIN_AREA,
    // The label in every language the plugin ships, for the settings search.
    keywords: Object.values(s.label),
    adminOnly: s.adminOnly,
    plugin: true,
    load: () => Promise.resolve((locked() ? LockedFeature : PluginHost) as never),
    props: () =>
      locked()
        ? { title: pick(s.label, locale()), description: pick(s.description, locale()) }
        : { mountable: s.view, context, props: {} },
  };
}

function toSearchAction(plugin: string, a: PluginSearchAction, context: PluginContext): SearchAction {
  return {
    id: `plugin-${plugin}-${a.id}`,
    keywords: a.keywords.map((k) => k.trim().toLocaleLowerCase()).filter((k) => k !== ''),
    title: () => pick(a.title, locale()),
    detail: () => (a.detail ? pick(a.detail, locale()) : ''),
    run: async () => {
      await a.run(context);
      return pick(a.done ?? a.title, locale());
    },
  };
}

/**
 * Whether the account may use a search action. Unlike settings sections,
 * an action that would only answer 403 is left out entirely: guests never
 * act, and a feature has to be fully licensed (read-only means no changes).
 */
const mayRun = (a: PluginSearchAction, context: PluginContext) =>
  !context.account.guest && (!a.feature || context.level(a.feature) === 'full');

/**
 * Whether the account gets a plugin's first-start step: never guests, only
 * the owner for owner-only steps, and only with the feature fully licensed,
 * so a step never saves something that is stored but not shown.
 */
const mayOnboard = (s: PluginOnboardingStep, context: PluginContext) =>
  !context.account.guest &&
  (!s.ownerOnly || context.account.owner) &&
  (!s.feature || context.level(s.feature) === 'full');

/**
 * Loads the plugins the server offers (GET /api/plugins) and registers their
 * widgets and settings sections. Without ee/ the list is empty and nothing
 * is downloaded. A plugin that fails to load is skipped; the app works on.
 */
export async function loadPlugins(
  http: HttpClient,
  auth: AuthUseCases,
  targets: PluginTargets,
  importPlugin: ImportPlugin = browserImport,
): Promise<string[]> {
  const { plugins } = await optional<{ plugins: string[] }>(http, '/api/plugins', { plugins: [] });
  if (!plugins?.length) return [];
  const { features } = await optional<{ features: Record<string, FeatureLevel> }>(
    http,
    '/api/licence/features',
    {
      features: {},
    },
  );
  const context: PluginContext = {
    http,
    get language() {
      return locale();
    },
    account: { owner: auth.owner, admin: auth.admin, member: auth.member, guest: auth.guest },
    level: (feature) => features[feature] ?? 'none',
    errorMessage,
  };
  const loaded: string[] = [];
  for (const url of plugins) {
    // Only plugins from this server: a relative path, never another origin.
    if (!url.startsWith('/') || url.startsWith('//')) continue;
    try {
      const { plugin } = await importPlugin(url);
      if (!plugin || !supportedVersion(plugin.apiVersion)) {
        console.warn('plugin skipped: unsupported API version', url);
        continue;
      }
      for (const w of plugin.widgets) {
        if (!targets.widgets.some((d) => d.type.kind === w.kind)) targets.widgets.push(toWidget(w, context));
      }
      for (const s of plugin.settings) targets.sections.push(toSection(s, context));
      // searchActions is optional: plugins built before it existed have none.
      for (const a of plugin.searchActions ?? []) {
        if (targets.searchActions && mayRun(a, context))
          targets.searchActions.push(toSearchAction(plugin.name, a, context));
      }
      for (const s of plugin.onboardingSteps ?? []) {
        if (targets.onboardingSteps && mayOnboard(s, context))
          targets.onboardingSteps.push({
            id: `plugin-${plugin.name}-${s.id}`,
            title: () => pick(s.title, locale()),
            mountable: s.view,
            context,
          });
      }
      for (const [language, texts] of Object.entries(plugin.serverMessages ?? {})) {
        registerServerMessages(language, texts);
      }
      loaded.push(plugin.name);
    } catch (err) {
      console.error('plugin failed to load', url, err);
    }
  }
  return loaded;
}
