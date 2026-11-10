/**
 * Widget registry (Strategy pattern): maps every widget kind to its
 * component, settings form and default size. Features never know about each
 * other; this file is the only place that combines them.
 */
import type { Component } from 'svelte';
import { BookmarksWidget, BookmarksWidgetSettings } from '../features/bookmarks';
import { ClockSettings, ClockWidget } from '../features/clock';
import type { WidgetType } from '../features/dashboard';
import { NotesWidget, NotesWidgetSettings } from '../features/notes';
import { RssWidget, RssWidgetSettings } from '../features/rss';
import { SearchWidget, SearchWidgetSettings } from '../features/search';
import { ServicesWidget } from '../features/services';
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
];

/** Looks up the definition of a widget kind. */
export function definitionFor(kind: string): WidgetDefinition | undefined {
  return widgetDefinitions.find((d) => d.type.kind === kind);
}
