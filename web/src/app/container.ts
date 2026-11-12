/**
 * Composition root of the frontend: creates the HTTP client, the API adapters
 * and the use cases, and hands them to the UI. Nothing else instantiates
 * adapters, so features stay independent of each other.
 */
import { createRefreshScheduler, type RefreshScheduler } from '../shared/refresh';
import { ShortcutRegistry } from '../shared/shortcuts/registry';
import { QuickActionsUseCases } from '../features/quickactions/usecase/quickActionsUseCases';
import { quickTargets } from './quickTargets';
import { AdGuardApi } from '../features/adguard/adapter/api/adguardApi';
import { AdGuardUseCases } from '../features/adguard/usecase/adguardUseCases';
import { FritzBoxApi } from '../features/fritzbox/adapter/api/fritzboxApi';
import { FritzBoxUseCases } from '../features/fritzbox/usecase/fritzboxUseCases';
import { NextcloudApi } from '../features/nextcloud/adapter/api/nextcloudApi';
import { NextcloudUseCases } from '../features/nextcloud/usecase/nextcloudUseCases';
import { ShoppingApi } from '../features/shopping/adapter/api/shoppingApi';
import { ShoppingUseCases } from '../features/shopping/usecase/shoppingUseCases';
import { ParcelsApi } from '../features/parcels/adapter/api/parcelsApi';
import { ParcelsUseCases } from '../features/parcels/usecase/parcelsUseCases';
import { TransferApi } from '../features/transfer/adapter/api/transferApi';
import { TransferUseCases } from '../features/transfer/usecase/transferUseCases';
import { TimelineApi } from '../features/timeline/adapter/api/timelineApi';
import { DiskHealthApi } from '../features/diskhealth/adapter/api/diskHealthApi';
import { SpeedtestApi } from '../features/speedtest/adapter/api/speedtestApi';
import { SpeedtestUseCases } from '../features/speedtest/usecase/speedtestUseCases';
import { DiskHealthUseCases } from '../features/diskhealth/usecase/diskHealthUseCases';
import { TimelineUseCases } from '../features/timeline/usecase/timelineUseCases';
import { TransitApi } from '../features/transit/adapter/api/transitApi';
import { TransitUseCases } from '../features/transit/usecase/transitUseCases';
import { WasteApi } from '../features/waste/adapter/api/wasteApi';
import { WasteUseCases } from '../features/waste/usecase/wasteUseCases';
import { WireGuardApi } from '../features/wireguard/adapter/api/wireguardApi';
import { WireGuardUseCases } from '../features/wireguard/usecase/wireguardUseCases';
import { AppearanceApi } from '../features/appearance/adapter/api/appearanceApi';
import { AppearanceUseCases } from '../features/appearance/usecase/appearanceUseCases';
import { AuthApi } from '../features/auth/adapter/api/authApi';
import { AuthUseCases } from '../features/auth/usecase/authUseCases';
import { BrowserPasskeys } from '../features/auth/adapter/browser/webauthn';
import { BackupApi } from '../features/backup/adapter/api/backupApi';
import { BackupUseCases } from '../features/backup/usecase/backupUseCases';
import { BackupMonitorApi } from '../features/backupmonitor/adapter/api/backupMonitorApi';
import { BackupMonitorUseCases } from '../features/backupmonitor/usecase/backupMonitorUseCases';
import { BookmarksApi } from '../features/bookmarks/adapter/api/bookmarksApi';
import { BookmarksUseCases } from '../features/bookmarks/usecase/bookmarksUseCases';
import { CalendarApi } from '../features/calendar/adapter/api/calendarApi';
import { CalendarUseCases } from '../features/calendar/usecase/calendarUseCases';
import { DashboardApi } from '../features/dashboard/adapter/api/dashboardApi';
import { DashboardUseCases } from '../features/dashboard/usecase/dashboardUseCases';
import { DockerApi } from '../features/docker/adapter/api/dockerApi';
import { DockerUseCases } from '../features/docker/usecase/dockerUseCases';
import { NotesApi } from '../features/notes/adapter/api/notesApi';
import { NotesUseCases } from '../features/notes/usecase/notesUseCases';
import { NotificationsApi } from '../features/notifications/adapter/api/notificationsApi';
import { NotificationsUseCases } from '../features/notifications/usecase/notificationsUseCases';
import { OnboardingApi } from '../features/onboarding/adapter/api/onboardingApi';
import { OnboardingUseCases } from '../features/onboarding/usecase/onboardingUseCases';
import { RssApi } from '../features/rss/adapter/api/rssApi';
import { RssUseCases } from '../features/rss/usecase/rssUseCases';
import { SearchApi } from '../features/search/adapter/api/searchApi';
import { SearchUseCases } from '../features/search/usecase/searchUseCases';
import { ServicesApi } from '../features/services/adapter/api/servicesApi';
import { ServicesUseCases } from '../features/services/usecase/servicesUseCases';
import { SystemApi } from '../features/system/adapter/api/systemApi';
import { SystemUseCases } from '../features/system/usecase/systemUseCases';
import { TodosApi } from '../features/todos/adapter/api/todosApi';
import { TodosUseCases } from '../features/todos/usecase/todosUseCases';
import { WeatherApi } from '../features/weather/adapter/api/weatherApi';
import { WeatherUseCases } from '../features/weather/usecase/weatherUseCases';
import type { InfoResponse } from '../shared/api-client/generated/modules';
import { FetchHttpClient } from '../shared/api-client/http';
import { BriefingApi } from '../features/briefing/adapter/api/briefingApi';
import { BriefingUseCases } from '../features/briefing/usecase/briefingUseCases';
import { CountdownApi } from '../features/countdown/adapter/api/countdownApi';
import { CountdownUseCases } from '../features/countdown/usecase/countdownUseCases';

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
  fritzbox: FritzBoxUseCases;
  wireguard: WireGuardUseCases;
  waste: WasteUseCases;
  shopping: ShoppingUseCases;
  transit: TransitUseCases;
  parcels: ParcelsUseCases;
  transfer: TransferUseCases;
  timeline: TimelineUseCases;
  diskhealth: DiskHealthUseCases;
  speedtest: SpeedtestUseCases;
  /** Shared refresh scheduler of all widgets (#117). */
  refresh: RefreshScheduler;
  /** Global keyboard shortcuts (#95). */
  shortcuts: ShortcutRegistry;
  /** Quick capture and palette commands (#105, #67). */
  quickActions: QuickActionsUseCases;
  briefing: BriefingUseCases;
  countdown: CountdownUseCases;
  // scaffold:container-fields
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
  const c: Omit<Container, 'quickActions'> = {
    http,
    auth: new AuthUseCases(new AuthApi(http), http, new BrowserPasskeys()),
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
    fritzbox: new FritzBoxUseCases(new FritzBoxApi(http)),
    wireguard: new WireGuardUseCases(new WireGuardApi(http)),
    waste: new WasteUseCases(new WasteApi(http)),
    shopping: new ShoppingUseCases(new ShoppingApi(http)),
    transit: new TransitUseCases(new TransitApi(http)),
    parcels: new ParcelsUseCases(new ParcelsApi(http)),
    transfer: new TransferUseCases(new TransferApi(http)),
    timeline: new TimelineUseCases(new TimelineApi(http)),
    diskhealth: new DiskHealthUseCases(new DiskHealthApi(http)),
    speedtest: new SpeedtestUseCases(new SpeedtestApi(http)),
    refresh: createRefreshScheduler(),
    shortcuts: new ShortcutRegistry(),
    briefing: new BriefingUseCases(new BriefingApi(http)),
    countdown: new CountdownUseCases(new CountdownApi(http)),
    // scaffold:container-init
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
  return { ...c, quickActions: new QuickActionsUseCases(quickTargets(c)) };
}
