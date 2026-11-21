import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '../../shared/api-client/http';
import type { AuthUseCases } from '../../features/auth';
import type { SearchAction } from '../searchActions';
import type { SettingsSection } from '../settingsKit';
import type { WidgetDefinition } from '../widgetKit';
import { PLUGIN_API_VERSION, type WaylightPlugin } from './api';
import { loadPlugins, supportedVersion, type OnboardingStep } from './loader';
import { serverMessage, setLanguage } from '../../shared/i18n';

const auth = { owner: true, admin: true, member: false, guest: false } as AuthUseCases;
const mountable = { mount: vi.fn(() => () => undefined) };

const plugin: WaylightPlugin = {
  apiVersion: PLUGIN_API_VERSION,
  name: 'ee',
  widgets: [
    {
      kind: 'announcements',
      feature: 'announcements',
      category: 'everyday',
      title: { en: 'Announcements', de: 'Ankündigungen' },
      description: { en: 'Notices for everyone' },
      defaultSize: { w: 4, h: 4 },
      icon: 'not-an-icon',
      view: mountable,
    },
  ],
  settings: [
    {
      id: 'branding',
      feature: 'branding',
      label: { en: 'Branding' },
      description: { en: 'Own name and colours' },
      adminOnly: true,
      view: mountable,
    },
    { id: 'licence', label: { en: 'Features for organisations' }, description: { en: '' }, view: mountable },
  ],
};

function http(routes: Record<string, unknown>): HttpClient {
  return {
    get: vi.fn((path: string) =>
      path in routes ? Promise.resolve(routes[path]) : Promise.reject(new Error('404')),
    ),
  } as unknown as HttpClient;
}

describe('loadPlugins', () => {
  it('loads nothing when the server offers no plugin (a build without ee/)', async () => {
    const targets = { widgets: [] as WidgetDefinition[], sections: [] as SettingsSection[] };
    const importPlugin = vi.fn();
    expect(await loadPlugins(http({}), auth, targets, importPlugin)).toEqual([]);
    expect(importPlugin).not.toHaveBeenCalled();
    expect(targets.widgets).toEqual([]);
  });

  it('registers widgets and sections; a section without a licence shows the locked view', async () => {
    const targets = { widgets: [] as WidgetDefinition[], sections: [] as SettingsSection[] };
    const routes = {
      '/api/plugins': { plugins: ['/ee/plugin.js', 'https://evil.example/x.js'] },
      '/api/licence/features': { features: { announcements: 'full', branding: 'none' } },
    };
    const importPlugin = vi.fn(() => Promise.resolve({ plugin }));
    expect(await loadPlugins(http(routes), auth, targets, importPlugin)).toEqual(['ee']);
    expect(importPlugin).toHaveBeenCalledTimes(1);
    expect(importPlugin).toHaveBeenCalledWith('/ee/plugin.js');

    const [widget] = targets.widgets;
    expect(widget?.type.kind).toBe('announcements');
    expect(widget?.type.title).toBe('Announcements');
    expect(widget?.icon).toBe('box');

    const [branding, licence] = targets.sections;
    expect(branding?.id).toBe('plugin-branding');
    expect(branding?.title?.()).toBe('Branding');
    expect(branding?.adminOnly).toBe(true);
    expect(branding?.props({} as never, { modules: new Set(), advanced: false })).toEqual({
      title: 'Branding',
      description: 'Own name and colours',
    });
    // A section without a feature is always open.
    expect(licence?.props({} as never, { modules: new Set(), advanced: false })).toMatchObject({ mountable });
  });

  it('skips a plugin built for another API version', async () => {
    const targets = { widgets: [] as WidgetDefinition[], sections: [] as SettingsSection[] };
    const old = { ...plugin, apiVersion: 99 };
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => undefined);
    const routes = { '/api/plugins': { plugins: ['/ee/plugin.js'] } };
    expect(await loadPlugins(http(routes), auth, targets, () => Promise.resolve({ plugin: old }))).toEqual(
      [],
    );
    expect(targets.widgets).toEqual([]);
    warn.mockRestore();
  });

  describe('search actions (#453)', () => {
    const run = vi.fn(() => Promise.resolve());
    const withActions: WaylightPlugin = {
      ...plugin,
      searchActions: [
        {
          id: 'office',
          feature: 'presence',
          keywords: [' Office ', 'Büro'],
          title: { en: 'Set office' },
          done: { en: 'Status set' },
          run,
        },
        { id: 'readonly', feature: 'branding', keywords: ['x'], title: { en: 'Read only' }, run },
        { id: 'free', keywords: ['free'], title: { en: 'Free' }, detail: { en: 'Core' }, run },
      ],
    };
    const routes = {
      '/api/plugins': { plugins: ['/ee/plugin.js'] },
      '/api/licence/features': { features: { presence: 'full', branding: 'read' } },
    };
    const load = async (account: AuthUseCases, p: WaylightPlugin = withActions) => {
      const targets = { widgets: [], sections: [], searchActions: [] as SearchAction[] };
      await loadPlugins(http(routes), account, targets, () => Promise.resolve({ plugin: p }));
      return targets.searchActions;
    };

    it('registers the actions of fully licensed features with normalised keywords', async () => {
      const actions = await load(auth);
      expect(actions.map((a) => a.id)).toEqual(['plugin-ee-office', 'plugin-ee-free']);
      expect(actions[0]?.keywords).toEqual(['office', 'büro']);
      expect(actions[1]?.detail()).toBe('Core');
      expect(await actions[1]?.run()).toBe('Free');
      expect(await actions[0]?.run()).toBe('Status set');
      expect(run).toHaveBeenCalledWith(expect.objectContaining({ account: expect.anything() }));
    });

    it('gives guests no actions', async () => {
      expect(await load({ ...auth, guest: true } as AuthUseCases)).toEqual([]);
    });

    it('still loads a plugin built before search actions existed', async () => {
      expect(await load(auth, plugin)).toEqual([]);
    });
  });
});

describe('first-start steps of plugins (#671)', () => {
  const withSteps: WaylightPlugin = {
    ...plugin,
    onboardingSteps: [
      {
        id: 'name',
        feature: 'branding',
        ownerOnly: true,
        title: { en: 'Name', de: 'Name' },
        view: mountable,
      },
      { id: 'free', title: { en: 'Free' }, view: mountable },
      { id: 'unlicensed', feature: 'presence', title: { en: 'Locked' }, view: mountable },
    ],
  };
  const routes = {
    '/api/plugins': { plugins: ['/ee/plugin.js'] },
    '/api/licence/features': { features: { branding: 'full', presence: 'read' } },
  };
  const load = async (account: AuthUseCases, p: WaylightPlugin = withSteps) => {
    const targets = { widgets: [], sections: [], onboardingSteps: [] as OnboardingStep[] };
    await loadPlugins(http(routes), account, targets, () => Promise.resolve({ plugin: p }));
    return targets.onboardingSteps.map((s) => s.id);
  };

  it('offers the owner the steps of fully licensed features', async () => {
    expect(await load(auth)).toEqual(['plugin-ee-name', 'plugin-ee-free']);
  });

  it('keeps owner-only steps from members and every step from guests', async () => {
    expect(await load({ ...auth, owner: false } as AuthUseCases)).toEqual(['plugin-ee-free']);
    expect(await load({ ...auth, guest: true } as AuthUseCases)).toEqual([]);
  });

  it('still loads a plugin built before the steps existed', async () => {
    expect(await load(auth, plugin)).toEqual([]);
  });
});

describe('compact views and backend texts of plugins (#632)', () => {
  const routes = { '/api/plugins': { plugins: ['/ee/plugin.js'] } };
  const withCompact: WaylightPlugin = {
    ...plugin,
    widgets: [
      {
        kind: 'presence',
        feature: 'presence',
        category: 'everyday',
        title: { en: 'Presence' },
        description: { en: '' },
        defaultSize: { w: 4, h: 4 },
        icon: 'activity',
        view: mountable,
        compact: (config) =>
          config['group'] ? { module: 'presence', params: { group: String(config['group']) } } : undefined,
        compactExample: { en: '3 in the office', de: '3 im Büro' },
      },
    ],
    serverMessages: { de: { srv_plugin_test: 'Plugin-Text {n}' } },
  };

  it('maps the compact source and registers the translations', async () => {
    const targets = { widgets: [] as WidgetDefinition[], sections: [] as SettingsSection[] };
    await loadPlugins(http(routes), auth, targets, () => Promise.resolve({ plugin: withCompact }));
    const widget = targets.widgets[0];
    expect(widget?.compact?.({ group: 'team a' })).toBe('presence?group=team+a');
    expect(widget?.compact?.({})).toBeUndefined();
    expect(widget?.compactExample?.()).toBe('3 in the office');
    setLanguage('de');
    expect(serverMessage('srv_plugin_test', { n: '2' })).toBe('Plugin-Text 2');
    setLanguage('en');
  });

  it('loads plugins of every supported version and no others', () => {
    expect(supportedVersion(1)).toBe(true);
    expect(supportedVersion(PLUGIN_API_VERSION)).toBe(true);
    expect(supportedVersion(0)).toBe(false);
    expect(supportedVersion(PLUGIN_API_VERSION + 1)).toBe(false);
    expect(supportedVersion('2')).toBe(false);
  });
});
