import { t } from '../shared/i18n';
/**
 * Widget registry (Strategy pattern): maps every widget kind to its
 * component, settings form and default size. Features never know about each
 * other; this file is the only place that combines them.
 */
import type { Component } from 'svelte';
import type { WidgetType } from '../features/dashboard';
import type { IconName } from '../shared/ui/icons/paths';
import type { Container } from './container';
import { prefixesFromConfig } from '../features/quickactions/domain/command';
import { quickCapture } from './quickCapture';

/** Configuration object stored per widget. */
type Config = Record<string, unknown>;

/** A lazily loaded component together with a function that builds its props (#116). */
export interface Renderer {
  /** Loads the component's chunk; the promise is cached per kind. */
  load: () => Promise<Component<Record<string, unknown>>>;
  props: (config: Config, c: Container) => Record<string, unknown>;
}

/** Everything the dashboard needs to know about one widget kind. */
export interface WidgetDefinition {
  type: WidgetType;
  /** Tile icon of the widget picker (#104); required for every kind. */
  icon: IconName;
  /** Rendered without card chrome outside edit mode. */
  chromeless?: boolean;
  view: Renderer;
  /** Settings form; it additionally receives an `onsave(config)` prop. */
  settings?: Renderer;
}

/**
 * Type-checks props against the lazily imported component and erases the
 * type for the registry. Only this file knows how widget code is loaded.
 */
function lazy<P extends object>(
  load: () => Promise<{ default: Component<P> }>,
  props: (config: Config, c: Container) => Omit<P, 'onsave'>,
): Renderer {
  let loaded: Promise<Component<Record<string, unknown>>> | undefined;
  return {
    load: () => {
      loaded ??= load().then((m) => m.default as unknown as Component<Record<string, unknown>>);
      // A failed chunk (e.g. offline after an update) can be retried.
      loaded.catch(() => (loaded = undefined));
      return loaded;
    },
    props,
  };
}

/** All widget kinds known to the frontend. */
export const widgetDefinitions: WidgetDefinition[] = [
  {
    type: {
      kind: 'search',
      category: 'everyday',
      keywords: ['web', 'google', 'duckduckgo', 'suche', 'shortcut'],
      get title() {
        return t('widget.search.title');
      },
      get description() {
        return t('widget.search.description');
      },
      defaultSize: { w: 8, h: 2 },
    },
    chromeless: true,
    icon: 'search',
    view: lazy(
      () => import('../features/search/ui/SearchWidget.svelte'),
      (config, c) => ({
        config,
        search: c.search,
        quick: quickCapture(c.quickActions, prefixesFromConfig(config['quickPrefixes'])),
      }),
    ),
    settings: lazy(
      () => import('../features/search/ui/SearchWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
  {
    type: {
      kind: 'clock',
      category: 'everyday',
      keywords: ['time', 'date', 'zeit', 'uhr', 'datum'],
      get title() {
        return t('widget.clock.title');
      },
      get description() {
        return t('widget.clock.description');
      },
      defaultSize: { w: 3, h: 3 },
    },
    icon: 'clock',
    view: lazy(
      () => import('../features/clock/ui/ClockWidget.svelte'),
      (config) => ({ config }),
    ),
    settings: lazy(
      () => import('../features/clock/ui/ClockSettings.svelte'),
      (config) => ({ config }),
    ),
  },
  {
    type: {
      kind: 'weather',
      category: 'information',
      keywords: ['forecast', 'temperature', 'wetter', 'vorhersage', 'regen'],
      get title() {
        return t('widget.weather.title');
      },
      get description() {
        return t('widget.weather.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'sun',
    view: lazy(
      () => import('../features/weather/ui/WeatherWidget.svelte'),
      (config, c) => ({ config, weather: c.weather }),
    ),
    settings: lazy(
      () => import('../features/weather/ui/WeatherSettings.svelte'),
      (config, c) => ({ config, weather: c.weather }),
    ),
  },
  {
    type: {
      kind: 'radar',
      category: 'information',
      keywords: ['rain', 'dwd', 'regenradar', 'niederschlag'],
      get title() {
        return t('widget.radar.title');
      },
      get description() {
        return t('widget.radar.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'rain',
    view: lazy(
      () => import('../features/weather/ui/RadarWidget.svelte'),
      (config, c) => ({ config, weather: c.weather }),
    ),
    settings: lazy(
      () => import('../features/weather/ui/WeatherSettings.svelte'),
      (config, c) => ({
        config,
        weather: c.weather,
        mode: 'radar' as const,
      }),
    ),
  },
  {
    type: {
      kind: 'services',
      category: 'server',
      keywords: ['uptime', 'status', 'health', 'dienste', 'erreichbarkeit'],
      get title() {
        return t('widget.services.title');
      },
      get description() {
        return t('widget.services.description');
      },
      defaultSize: { w: 6, h: 4 },
    },
    icon: 'server',
    view: lazy(
      () => import('../features/services/ui/ServicesWidget.svelte'),
      (config, c) => ({ config, services: c.services }),
    ),
  },
  {
    type: {
      kind: 'status-overview',
      category: 'server',
      keywords: ['uptime', 'incidents', 'sla', 'statusseite', 'störungen'],
      get title() {
        return t('widget.status-overview.title');
      },
      get description() {
        return t('widget.status-overview.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'activity',
    view: lazy(
      () => import('../features/services/ui/StatusOverviewWidget.svelte'),
      (_config, c) => ({ services: c.services }),
    ),
  },
  {
    type: {
      kind: 'bookmarks',
      category: 'everyday',
      keywords: ['links', 'favorites', 'lesezeichen', 'favoriten'],
      get title() {
        return t('widget.bookmarks.title');
      },
      get description() {
        return t('widget.bookmarks.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'bookmark',
    view: lazy(
      () => import('../features/bookmarks/ui/BookmarksWidget.svelte'),
      (config, c) => ({ config, bookmarks: c.bookmarks }),
    ),
    settings: lazy(
      () => import('../features/bookmarks/ui/BookmarksWidgetSettings.svelte'),
      (config, c) => ({ config, bookmarks: c.bookmarks }),
    ),
  },
  {
    type: {
      kind: 'notes',
      category: 'everyday',
      keywords: ['markdown', 'text', 'notizen', 'memo'],
      get title() {
        return t('widget.notes.title');
      },
      get description() {
        return t('widget.notes.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'note',
    view: lazy(
      () => import('../features/notes/ui/NotesWidget.svelte'),
      (config, c) => ({ config, notes: c.notes }),
    ),
    settings: lazy(
      () => import('../features/notes/ui/NotesWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
  {
    type: {
      kind: 'todos',
      category: 'everyday',
      keywords: ['tasks', 'to-do', 'aufgaben', 'erledigen'],
      get title() {
        return t('widget.todos.title');
      },
      get description() {
        return t('widget.todos.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'check',
    view: lazy(
      () => import('../features/todos/ui/TodosWidget.svelte'),
      (config, c) => ({ config, todos: c.todos }),
    ),
    settings: lazy(
      () => import('../features/todos/ui/TodosWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
  {
    type: {
      kind: 'rss',
      category: 'information',
      keywords: ['news', 'feed', 'atom', 'nachrichten'],
      get title() {
        return t('widget.rss.title');
      },
      get description() {
        return t('widget.rss.description');
      },
      defaultSize: { w: 5, h: 6 },
    },
    icon: 'rss',
    view: lazy(
      () => import('../features/rss/ui/RssWidget.svelte'),
      (config, c) => ({ config, rss: c.rss }),
    ),
    settings: lazy(
      () => import('../features/rss/ui/RssWidgetSettings.svelte'),
      (config, c) => ({ config, rss: c.rss }),
    ),
  },
  {
    type: {
      kind: 'calendar',
      category: 'information',
      keywords: ['events', 'caldav', 'ical', 'kalender', 'termine'],
      get title() {
        return t('widget.calendar.title');
      },
      get description() {
        return t('widget.calendar.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'calendar',
    view: lazy(
      () => import('../features/calendar/ui/CalendarWidget.svelte'),
      (config, c) => ({ config, calendar: c.calendar }),
    ),
    settings: lazy(
      () => import('../features/calendar/ui/CalendarWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
  {
    type: {
      kind: 'system',
      category: 'server',
      keywords: ['cpu', 'memory', 'ram', 'disk', 'host', 'speicher'],
      get title() {
        return t('widget.system.title');
      },
      get description() {
        return t('widget.system.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'cpu',
    view: lazy(
      () => import('../features/system/ui/SystemWidget.svelte'),
      (_config, c) => ({ system: c.system }),
    ),
  },
  {
    type: {
      kind: 'docker',
      category: 'server',
      keywords: ['containers', 'compose', 'images', 'container'],
      get title() {
        return t('widget.docker.title');
      },
      get description() {
        return t('widget.docker.description');
      },
      defaultSize: { w: 4, h: 6 },
    },
    icon: 'box',
    view: lazy(
      () => import('../features/docker/ui/DockerWidget.svelte'),
      (config, c) => ({ config, docker: c.docker }),
    ),
    settings: lazy(
      () => import('../features/docker/ui/DockerWidgetSettings.svelte'),
      (config, c) => ({ config, docker: c.docker }),
    ),
  },
  {
    type: {
      kind: 'adguard',
      category: 'network',
      keywords: ['dns', 'ads', 'blocking', 'werbung'],
      get title() {
        return t('widget.adguard.title');
      },
      get description() {
        return t('widget.adguard.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'shield',
    view: lazy(
      () => import('../features/adguard/ui/AdGuardWidget.svelte'),
      (_config, c) => ({ adguard: c.adguard }),
    ),
  },
  {
    type: {
      kind: 'nextcloud',
      category: 'network',
      keywords: ['files', 'storage', 'dateien', 'cloud'],
      get title() {
        return t('widget.nextcloud.title');
      },
      get description() {
        return t('widget.nextcloud.description');
      },
      defaultSize: { w: 4, h: 3 },
    },
    icon: 'cloud',
    view: lazy(
      () => import('../features/nextcloud/ui/NextcloudWidget.svelte'),
      (_config, c) => ({ nextcloud: c.nextcloud }),
    ),
  },
  {
    type: {
      kind: 'fritzbox',
      category: 'network',
      keywords: ['avm', 'router', 'wifi', 'wlan', 'internet'],
      get title() {
        return t('widget.fritzbox.title');
      },
      get description() {
        return t('widget.fritzbox.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'router',
    view: lazy(
      () => import('../features/fritzbox/ui/FritzBoxWidget.svelte'),
      (_config, c) => ({ fritzbox: c.fritzbox }),
    ),
  },
  {
    type: {
      kind: 'wireguard',
      category: 'network',
      keywords: ['vpn', 'peers', 'tunnel'],
      get title() {
        return t('widget.wireguard.title');
      },
      get description() {
        return t('widget.wireguard.description');
      },
      defaultSize: { w: 4, h: 3 },
    },
    icon: 'key',
    view: lazy(
      () => import('../features/wireguard/ui/WireGuardWidget.svelte'),
      (_config, c) => ({ wireguard: c.wireguard }),
    ),
  },
  {
    type: {
      kind: 'waste',
      category: 'information',
      keywords: ['garbage', 'bin', 'müll', 'abfall', 'tonne'],
      get title() {
        return t('widget.waste.title');
      },
      get description() {
        return t('widget.waste.description');
      },
      defaultSize: { w: 3, h: 3 },
    },
    icon: 'trash',
    view: lazy(
      () => import('../features/waste/ui/WasteWidget.svelte'),
      (_config, c) => ({ waste: c.waste }),
    ),
  },
  {
    type: {
      kind: 'shopping',
      category: 'information',
      keywords: ['groceries', 'list', 'einkaufen', 'einkaufsliste'],
      get title() {
        return t('widget.shopping.title');
      },
      get description() {
        return t('widget.shopping.description');
      },
      defaultSize: { w: 3, h: 4 },
    },
    icon: 'cart',
    view: lazy(
      () => import('../features/shopping/ui/ShoppingWidget.svelte'),
      (_config, c) => ({ shopping: c.shopping }),
    ),
  },
  {
    type: {
      kind: 'transit',
      category: 'information',
      keywords: ['departures', 'train', 'abfahrten', 'bahn', 'öpnv'],
      get title() {
        return t('widget.transit.title');
      },
      get description() {
        return t('widget.transit.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'bus',
    view: lazy(
      () => import('../features/transit/ui/TransitWidget.svelte'),
      (config, c) => ({ config, transit: c.transit }),
    ),
    settings: lazy(
      () => import('../features/transit/ui/TransitWidgetSettings.svelte'),
      (config, c) => ({ config, transit: c.transit }),
    ),
  },
  {
    type: {
      kind: 'parcels',
      category: 'information',
      keywords: ['dhl', 'tracking', 'pakete', 'sendung'],
      get title() {
        return t('widget.parcels.title');
      },
      get description() {
        return t('widget.parcels.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'package',
    view: lazy(
      () => import('../features/parcels/ui/ParcelsWidget.svelte'),
      (_config, c) => ({ parcels: c.parcels }),
    ),
  },
  {
    type: {
      kind: 'backup-status',
      category: 'server',
      keywords: ['restic', 'borg', 'cron', 'heartbeat', 'sicherung'],
      get title() {
        return t('widget.backup-status.title');
      },
      get description() {
        return t('widget.backup-status.description');
      },
      defaultSize: { w: 4, h: 3 },
    },
    icon: 'archive',
    view: lazy(
      () => import('../features/backupmonitor/ui/BackupStatusWidget.svelte'),
      (_config, c) => ({ backupMonitor: c.backupMonitor }),
    ),
  },
  {
    type: {
      kind: 'timeline',
      category: 'information',
      keywords: ['events', 'log', 'history', 'zeitleiste', 'verlauf'],
      get title() {
        return t('widget.timeline.title');
      },
      get description() {
        return t('widget.timeline.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'history',
    view: lazy(
      () => import('../features/timeline/ui/TimelineWidget.svelte'),
      (config, c) => ({ config, timeline: c.timeline }),
    ),
  },
  {
    type: {
      kind: 'diskhealth',
      category: 'server',
      keywords: ['smart', 'nvme', 'hdd', 'ssd', 'festplatte'],
      get title() {
        return t('widget.diskhealth.title');
      },
      get description() {
        return t('widget.diskhealth.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'disk',
    view: lazy(
      () => import('../features/diskhealth/ui/DiskHealthWidget.svelte'),
      (_config, c) => ({ diskhealth: c.diskhealth }),
    ),
  },
  {
    type: {
      kind: 'speedtest',
      category: 'server',
      keywords: ['internet', 'bandwidth', 'ping', 'bandbreite'],
      get title() {
        return t('widget.speedtest.title');
      },
      get description() {
        return t('widget.speedtest.description');
      },
      defaultSize: { w: 3, h: 4 },
    },
    icon: 'gauge',
    view: lazy(
      () => import('../features/speedtest/ui/SpeedtestWidget.svelte'),
      (_config, c) => ({ speedtest: c.speedtest }),
    ),
  },
  {
    type: {
      kind: 'arr',
      category: 'server',
      keywords: ['sonarr', 'radarr', 'prowlarr', 'downloads', 'series', 'movies'],
      get title() {
        return t('widget.arr.title');
      },
      get description() {
        return t('widget.arr.description');
      },
      defaultSize: { w: 4, h: 6 },
    },
    icon: 'package',
    view: lazy(
      () => import('../features/arr/ui/ArrWidget.svelte'),
      (config, c) => ({ config, arr: c.arr }),
    ),
    settings: lazy(
      () => import('../features/arr/ui/ArrWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
  // scaffold:widgets - make new-module inserts new widgets above this line.
];

/** Picker icon of a widget kind. */
export function iconFor(kind: string): IconName | undefined {
  return definitionFor(kind)?.icon;
}

/** Looks up the definition of a widget kind. */
export function definitionFor(kind: string): WidgetDefinition | undefined {
  return widgetDefinitions.find((d) => d.type.kind === kind);
}
