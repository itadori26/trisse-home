/**
 * Composition root of the frontend: creates the HTTP client, the API adapters
 * and the use cases, and hands them to the UI. Nothing else instantiates
 * adapters, so features stay independent of each other.
 */
import { AdGuardApi, AdGuardUseCases } from '../features/adguard';
import { NextcloudApi, NextcloudUseCases } from '../features/nextcloud';
import { AppearanceApi, AppearanceUseCases } from '../features/appearance';
import { AuthApi, AuthUseCases } from '../features/auth';
import { BackupApi, BackupUseCases } from '../features/backup';
import { BackupMonitorApi, BackupMonitorUseCases } from '../features/backupmonitor';
import { BookmarksApi, BookmarksUseCases } from '../features/bookmarks';
import { CalendarApi, CalendarUseCases } from '../features/calendar';
import { DashboardApi, DashboardUseCases } from '../features/dashboard';
import { DockerApi, DockerUseCases } from '../features/docker';
import { NotesApi, NotesUseCases } from '../features/notes';
import { NotificationsApi, NotificationsUseCases } from '../features/notifications';
import { OnboardingApi, OnboardingUseCases } from '../features/onboarding';
import { RssApi, RssUseCases } from '../features/rss';
import { SearchApi, SearchUseCases } from '../features/search';
import { ServicesApi, ServicesUseCases } from '../features/services';
import { SystemApi, SystemUseCases } from '../features/system';
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
  system: SystemUseCases;
  docker: DockerUseCases;
  adguard: AdGuardUseCases;
  nextcloud: NextcloudUseCases;
  calendar: CalendarUseCases;
  onboarding: OnboardingUseCases;
  backup: BackupUseCases;
  backupMonitor: BackupMonitorUseCases;
  notifications: NotificationsUseCases;
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
    system: new SystemUseCases(new SystemApi(http)),
    docker: new DockerUseCases(new DockerApi(http)),
    adguard: new AdGuardUseCases(new AdGuardApi(http)),
    nextcloud: new NextcloudUseCases(new NextcloudApi(http)),
    calendar: new CalendarUseCases(new CalendarApi(http)),
    onboarding: new OnboardingUseCases(new OnboardingApi(http)),
    backup: new BackupUseCases(new BackupApi(http)),
    backupMonitor: new BackupMonitorUseCases(new BackupMonitorApi(http)),
    notifications: new NotificationsUseCases(new NotificationsApi(http)),
    async enabledModules() {
      const info = await http.get<InfoResponse>('/api/modules');
      return { modules: new Set(info.modules), widgetKinds: new Set(info.widgetKinds) };
    },
  };
}
