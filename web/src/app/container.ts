/**
 * Composition root of the frontend: creates the HTTP client, the API adapters
 * and the use cases, and hands them to the UI. Nothing else instantiates
 * adapters, so features stay independent of each other.
 */
import { AppearanceApi, AppearanceUseCases } from '../features/appearance';
import { AuthApi, AuthUseCases } from '../features/auth';
import { BookmarksApi, BookmarksUseCases } from '../features/bookmarks';
import { DashboardApi, DashboardUseCases } from '../features/dashboard';
import { NotesApi, NotesUseCases } from '../features/notes';
import { RssApi, RssUseCases } from '../features/rss';
import { SearchApi, SearchUseCases } from '../features/search';
import { ServicesApi, ServicesUseCases } from '../features/services';
import { TodosApi, TodosUseCases } from '../features/todos';
import { WeatherApi, WeatherUseCases } from '../features/weather';
import type { InfoResponse } from '../shared/api-client/generated/modules';
import { FetchHttpClient } from '../shared/api-client/http';

/** All use cases available to the UI. */
export interface Container {
  http: FetchHttpClient;
  auth: AuthUseCases;
  dashboard: DashboardUseCases;
  appearance: AppearanceUseCases;
  search: SearchUseCases;
  weather: WeatherUseCases;
  services: ServicesUseCases;
  bookmarks: BookmarksUseCases;
  notes: NotesUseCases;
  todos: TodosUseCases;
  rss: RssUseCases;
  /** Loads which backend modules and widget kinds are enabled. */
  enabledModules(): Promise<{ modules: Set<string>; widgetKinds: Set<string> }>;
}

/** Builds the container. `onUnauthorized` fires when the session expired. */
export function createContainer(onUnauthorized: () => void): Container {
  const http = new FetchHttpClient(onUnauthorized);
  return {
    http,
    auth: new AuthUseCases(new AuthApi(http), http),
    dashboard: new DashboardUseCases(new DashboardApi(http)),
    appearance: new AppearanceUseCases(new AppearanceApi(http)),
    search: new SearchUseCases(new SearchApi(http)),
    weather: new WeatherUseCases(new WeatherApi(http)),
    services: new ServicesUseCases(new ServicesApi(http)),
    bookmarks: new BookmarksUseCases(new BookmarksApi(http)),
    notes: new NotesUseCases(new NotesApi(http)),
    todos: new TodosUseCases(new TodosApi(http)),
    rss: new RssUseCases(new RssApi(http)),
    async enabledModules() {
      const info = await http.get<InfoResponse>('/api/modules');
      return { modules: new Set(info.modules), widgetKinds: new Set(info.widgetKinds) };
    },
  };
}
