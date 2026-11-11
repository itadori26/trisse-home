/**
 * Widget registry (Strategy pattern): maps every widget kind to its
 * component, settings form and default size. Features never know about each
 * other; this file is the only place that combines them.
 */
import type { Component } from 'svelte';
import { AdGuardWidget } from '../features/adguard';
import { FritzBoxWidget } from '../features/fritzbox';
import { NextcloudWidget } from '../features/nextcloud';
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
import { ServicesWidget } from '../features/services';
import { SystemWidget } from '../features/system';
import { TodosWidget, TodosWidgetSettings } from '../features/todos';
import { WeatherSettings, WeatherWidget } from '../features/weather';
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
      title: 'Search',
      description: 'Search bar (DuckDuckGo by default, !bangs work).',
      defaultSize: { w: 8, h: 2 },
    },
    chromeless: true,
    view: renderer(SearchWidget, (config, c) => ({ config, search: c.search })),
    settings: renderer(SearchWidgetSettings, (config) => ({ config })),
  },
  {
    type: {
      kind: 'clock',
      title: 'Clock',
      description: 'Time and date, optionally in another time zone.',
      defaultSize: { w: 3, h: 3 },
    },
    view: renderer(ClockWidget, (config) => ({ config })),
    settings: renderer(ClockSettings, (config) => ({ config })),
  },
  {
    type: {
      kind: 'weather',
      title: 'Weather',
      description: 'Current weather and forecast (Open-Meteo).',
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(WeatherWidget, (config, c) => ({ config, weather: c.weather })),
    settings: renderer(WeatherSettings, (config, c) => ({ config, weather: c.weather })),
  },
  {
    type: {
      kind: 'services',
      title: 'Services',
      description: 'Tiles for your services with online status.',
      defaultSize: { w: 6, h: 4 },
    },
    view: renderer(ServicesWidget, (config, c) => ({ config, services: c.services })),
  },
  {
    type: {
      kind: 'bookmarks',
      title: 'Bookmarks',
      description: 'Links from a folder or with a tag.',
      defaultSize: { w: 4, h: 5 },
    },
    view: renderer(BookmarksWidget, (config, c) => ({ config, bookmarks: c.bookmarks })),
    settings: renderer(BookmarksWidgetSettings, (config, c) => ({ config, bookmarks: c.bookmarks })),
  },
  {
    type: {
      kind: 'notes',
      title: 'Notes',
      description: 'Pinned Markdown notes.',
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(NotesWidget, (config, c) => ({ config, notes: c.notes })),
    settings: renderer(NotesWidgetSettings, (config) => ({ config })),
  },
  {
    type: {
      kind: 'todos',
      title: 'To-dos',
      description: 'A to-do list with due dates.',
      defaultSize: { w: 4, h: 5 },
    },
    view: renderer(TodosWidget, (config, c) => ({ config, todos: c.todos })),
    settings: renderer(TodosWidgetSettings, (config) => ({ config })),
  },
  {
    type: {
      kind: 'rss',
      title: 'News',
      description: 'Latest items of your RSS/Atom feeds.',
      defaultSize: { w: 5, h: 6 },
    },
    view: renderer(RssWidget, (config, c) => ({ config, rss: c.rss })),
    settings: renderer(RssWidgetSettings, (config, c) => ({ config, rss: c.rss })),
  },
  {
    type: {
      kind: 'calendar',
      title: 'Calendar',
      description: 'Upcoming events from CalDAV (e.g. Nextcloud).',
      defaultSize: { w: 4, h: 5 },
    },
    view: renderer(CalendarWidget, (config, c) => ({ config, calendar: c.calendar })),
    settings: renderer(CalendarWidgetSettings, (config) => ({ config })),
  },
  {
    type: {
      kind: 'system',
      title: 'System',
      description: 'CPU, memory, disks, load and uptime of the host.',
      defaultSize: { w: 4, h: 5 },
    },
    view: renderer(SystemWidget, (_config, c) => ({ system: c.system })),
  },
  {
    type: {
      kind: 'docker',
      title: 'Docker',
      description: 'Container status via the read-only socket proxy.',
      defaultSize: { w: 4, h: 6 },
    },
    view: renderer(DockerWidget, (config, c) => ({ config, docker: c.docker })),
    settings: renderer(DockerWidgetSettings, (config, c) => ({ config, docker: c.docker })),
  },
  {
    type: {
      kind: 'adguard',
      title: 'AdGuard Home',
      description: 'DNS queries and blocking statistics.',
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(AdGuardWidget, (_config, c) => ({ adguard: c.adguard })),
  },
  {
    type: {
      kind: 'nextcloud',
      title: 'Nextcloud',
      description: 'Version, your quota and server status of your Nextcloud.',
      defaultSize: { w: 4, h: 3 },
    },
    view: renderer(NextcloudWidget, (_config, c) => ({ nextcloud: c.nextcloud })),
  },
  {
    type: {
      kind: 'fritzbox',
      title: 'FRITZ!Box',
      description: 'Internet connection, current throughput and recent calls.',
      defaultSize: { w: 4, h: 4 },
    },
    view: renderer(FritzBoxWidget, (_config, c) => ({ fritzbox: c.fritzbox })),
  },
  {
    type: {
      kind: 'wireguard',
      title: 'WireGuard',
      description: 'VPN peers with last handshake and traffic.',
      defaultSize: { w: 4, h: 3 },
    },
    view: renderer(WireGuardWidget, (_config, c) => ({ wireguard: c.wireguard })),
  },
  {
    type: {
      kind: 'waste',
      title: 'Waste collection',
      description: 'Next collection dates from your municipality’s ICS calendar.',
      defaultSize: { w: 3, h: 3 },
    },
    view: renderer(WasteWidget, (_config, c) => ({ waste: c.waste })),
  },
  {
    type: {
      kind: 'backup-status',
      title: 'Backups',
      description: 'Status of your monitored backup jobs.',
      defaultSize: { w: 4, h: 3 },
    },
    view: renderer(BackupStatusWidget, (_config, c) => ({ backupMonitor: c.backupMonitor })),
  },
];

/** Looks up the definition of a widget kind. */
export function definitionFor(kind: string): WidgetDefinition | undefined {
  return widgetDefinitions.find((d) => d.type.kind === kind);
}
