import { t } from '../shared/i18n';
/**
 * Widget registry (Strategy pattern): maps every widget kind to its
 * component, settings form and default size. Features never know about each
 * other; this file is the only place that combines them.
 */
import type { Component } from 'svelte';
import { AdGuardWidget } from '../features/adguard';
import { FritzBoxWidget } from '../features/fritzbox';
import { NextcloudWidget } from '../features/nextcloud';
import { ShoppingWidget } from '../features/shopping';
import { ParcelsWidget } from '../features/parcels';
import { TransitWidget, TransitWidgetSettings } from '../features/transit';
import { WasteWidget } from '../features/waste';
import { WireGuardWidget } from '../features/wireguard';
import { BackupStatusWidget } from '../features/backupmonitor';
import { BookmarksWidget, BookmarksWidgetSettings } from '../features/bookmarks';
import { CalendarWidget, CalendarWidgetSettings } from '../features/calendar';
import { ClockSettings, ClockWidget } from '../features/clock';
import type { WidgetType } from '../features/dashboard';
import { DockerWidget, DockerWidgetSettings } from '../features/docker';
import { NotesWidget, NotesWidgetSettings } from '../features/notes';
import { RssWidget, RssWidgetSettings } from '../features/rss';
import { SearchWidget, SearchWidgetSettings } from '../features/search';
import { ServicesWidget, StatusOverviewWidget } from '../features/services';
import { TimelineWidget } from '../features/timeline';
import { DiskHealthWidget } from '../features/diskhealth';
import { SystemWidget } from '../features/system';
import { TodosWidget, TodosWidgetSettings } from '../features/todos';
import { RadarWidget, WeatherSettings, WeatherWidget } from '../features/weather';
import type { Container } from './container';

/** Configuration object stored per widget. */
type Config = Record<string, unknown>;

/** A component together with a function that builds its props. */
interface Renderer {
  component: Component<Record<string, unknown>>;
  props: (config: Config, c: Container) => Record<string, unknown>;
}

/** Everything the dashboard needs to know about one widget kind. */
export interface WidgetDefinition {
  type: WidgetType;
  /** Rendered without card chrome outside edit mode. */
  chromeless?: boolean;
  view: Renderer;
  /** Settings form; it additionally receives an `onsave(config)` prop. */
  settings?: Renderer;
}

/** Type-checks props against the component and erases the type for the registry. */
function renderer<P extends object>(
  component: Component<P>,
  props: (config: Config, c: Container) => Omit<P, 'onsave'>,
): Renderer {
  return { component: component as unknown as Component<Record<string, unknown>>, props };
}

/** All widget kinds known to the frontend. */
export const widgetDefinitions: WidgetDefinition[] = [
  {
    type: {
      kind: 'search',
      get title() {
        return t('widget.search.title');
      },
      get description() {
        return t('widget.search.description');
      },
      defaultSize: { w: 8, h: 2 },
    },
    chromeless: true,
    view: renderer(SearchWidget, (config, c) => ({ config, search: c.search })),
    settings: renderer(SearchWidgetSettings, (config) => ({ config })),
  },
  {
    type: {
      kind: 'clock',
      get title() {
        return t('widget.clock.title');
      },
      get description() {
        return t('widget.clock.description');
      },
      defaultSize: { w: 3, h: 3 },
    },
    view: renderer(ClockWidget, (config) => ({ config })),
    settings: renderer(ClockSettings, (config) => ({ config })),
  },
  {
    type: {
      kind: 'weather',
      get title() {
        return t('widget.weather.title');
      },
      get description() {
        return t('widget.weather.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(WeatherWidget, (config, c) => ({ config, weather: c.weather })),
    settings: renderer(WeatherSettings, (config, c) => ({ config, weather: c.weather })),
  },
  {
    type: {
      kind: 'radar',
      get title() {
        return t('widget.radar.title');
      },
      get description() {
        return t('widget.radar.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(RadarWidget, (config, c) => ({ config, weather: c.weather })),
    settings: renderer(WeatherSettings, (config, c) => ({
      config,
      weather: c.weather,
      mode: 'radar' as const,
    })),
  },
  {
    type: {
      kind: 'services',
      get title() {
        return t('widget.services.title');
      },
      get description() {
        return t('widget.services.description');
      },
      defaultSize: { w: 6, h: 4 },
    },
    view: renderer(ServicesWidget, (config, c) => ({ config, services: c.services })),
  },
  {
    type: {
      kind: 'status-overview',
      get title() {
        return t('widget.status-overview.title');
      },
      get description() {
        return t('widget.status-overview.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    view: renderer(StatusOverviewWidget, (_config, c) => ({ services: c.services })),
  },
  {
    type: {
      kind: 'bookmarks',
      get title() {
        return t('widget.bookmarks.title');
      },
      get description() {
        return t('widget.bookmarks.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    view: renderer(BookmarksWidget, (config, c) => ({ config, bookmarks: c.bookmarks })),
    settings: renderer(BookmarksWidgetSettings, (config, c) => ({ config, bookmarks: c.bookmarks })),
  },
  {
    type: {
      kind: 'notes',
      get title() {
        return t('widget.notes.title');
      },
      get description() {
        return t('widget.notes.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(NotesWidget, (config, c) => ({ config, notes: c.notes })),
    settings: renderer(NotesWidgetSettings, (config) => ({ config })),
  },
  {
    type: {
      kind: 'todos',
      get title() {
        return t('widget.todos.title');
      },
      get description() {
        return t('widget.todos.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    view: renderer(TodosWidget, (config, c) => ({ config, todos: c.todos })),
    settings: renderer(TodosWidgetSettings, (config) => ({ config })),
  },
  {
    type: {
      kind: 'rss',
      get title() {
        return t('widget.rss.title');
      },
      get description() {
        return t('widget.rss.description');
      },
      defaultSize: { w: 5, h: 6 },
    },
    view: renderer(RssWidget, (config, c) => ({ config, rss: c.rss })),
    settings: renderer(RssWidgetSettings, (config, c) => ({ config, rss: c.rss })),
  },
  {
    type: {
      kind: 'calendar',
      get title() {
        return t('widget.calendar.title');
      },
      get description() {
        return t('widget.calendar.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    view: renderer(CalendarWidget, (config, c) => ({ config, calendar: c.calendar })),
    settings: renderer(CalendarWidgetSettings, (config) => ({ config })),
  },
  {
    type: {
      kind: 'system',
      get title() {
        return t('widget.system.title');
      },
      get description() {
        return t('widget.system.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    view: renderer(SystemWidget, (_config, c) => ({ system: c.system })),
  },
  {
    type: {
      kind: 'docker',
      get title() {
        return t('widget.docker.title');
      },
      get description() {
        return t('widget.docker.description');
      },
      defaultSize: { w: 4, h: 6 },
    },
    view: renderer(DockerWidget, (config, c) => ({ config, docker: c.docker })),
    settings: renderer(DockerWidgetSettings, (config, c) => ({ config, docker: c.docker })),
  },
  {
    type: {
      kind: 'adguard',
      get title() {
        return t('widget.adguard.title');
      },
      get description() {
        return t('widget.adguard.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(AdGuardWidget, (_config, c) => ({ adguard: c.adguard })),
  },
  {
    type: {
      kind: 'nextcloud',
      get title() {
        return t('widget.nextcloud.title');
      },
      get description() {
        return t('widget.nextcloud.description');
      },
      defaultSize: { w: 4, h: 3 },
    },
    view: renderer(NextcloudWidget, (_config, c) => ({ nextcloud: c.nextcloud })),
  },
  {
    type: {
      kind: 'fritzbox',
      get title() {
        return t('widget.fritzbox.title');
      },
      get description() {
        return t('widget.fritzbox.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(FritzBoxWidget, (_config, c) => ({ fritzbox: c.fritzbox })),
  },
  {
    type: {
      kind: 'wireguard',
      get title() {
        return t('widget.wireguard.title');
      },
      get description() {
        return t('widget.wireguard.description');
      },
      defaultSize: { w: 4, h: 3 },
    },
    view: renderer(WireGuardWidget, (_config, c) => ({ wireguard: c.wireguard })),
  },
  {
    type: {
      kind: 'waste',
      get title() {
        return t('widget.waste.title');
      },
      get description() {
        return t('widget.waste.description');
      },
      defaultSize: { w: 3, h: 3 },
    },
    view: renderer(WasteWidget, (_config, c) => ({ waste: c.waste })),
  },
  {
    type: {
      kind: 'shopping',
      get title() {
        return t('widget.shopping.title');
      },
      get description() {
        return t('widget.shopping.description');
      },
      defaultSize: { w: 3, h: 4 },
    },
    view: renderer(ShoppingWidget, (_config, c) => ({ shopping: c.shopping })),
  },
  {
    type: {
      kind: 'transit',
      get title() {
        return t('widget.transit.title');
      },
      get description() {
        return t('widget.transit.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(TransitWidget, (config, c) => ({ config, transit: c.transit })),
    settings: renderer(TransitWidgetSettings, (config, c) => ({ config, transit: c.transit })),
  },
  {
    type: {
      kind: 'parcels',
      get title() {
        return t('widget.parcels.title');
      },
      get description() {
        return t('widget.parcels.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(ParcelsWidget, (_config, c) => ({ parcels: c.parcels })),
  },
  {
    type: {
      kind: 'backup-status',
      get title() {
        return t('widget.backup-status.title');
      },
      get description() {
        return t('widget.backup-status.description');
      },
      defaultSize: { w: 4, h: 3 },
    },
    view: renderer(BackupStatusWidget, (_config, c) => ({ backupMonitor: c.backupMonitor })),
  },
  {
    type: {
      kind: 'timeline',
      get title() {
        return t('widget.timeline.title');
      },
      get description() {
        return t('widget.timeline.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    view: renderer(TimelineWidget, (config, c) => ({ config, timeline: c.timeline })),
  },
  {
    type: {
      kind: 'diskhealth',
      get title() {
        return t('widget.diskhealth.title');
      },
      get description() {
        return t('widget.diskhealth.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(DiskHealthWidget, (_config, c) => ({ diskhealth: c.diskhealth })),
  },
  // scaffold:widgets - make new-module inserts new widgets above this line.
];

/** Looks up the definition of a widget kind. */
export function definitionFor(kind: string): WidgetDefinition | undefined {
  return widgetDefinitions.find((d) => d.type.kind === kind);
}
