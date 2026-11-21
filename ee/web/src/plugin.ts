/**
 * Entry of the ee/ frontend (ADR 0021): what the web app registers when the
 * server offers this bundle. Features without a licence stay listed in the
 * settings; their widgets are hidden by the server's module list.
 */
import { PLUGIN_API_VERSION, type WaylightPlugin } from '../../../web/src/app/plugins/api';
import { texts as absenceTexts } from './absences/texts';
import { texts as announcementTexts } from './announcements/texts';
import { texts as brandingTexts } from './branding/texts';
import { texts as connectionTexts } from './connections/texts';
import { texts as teamCalendarTexts } from './teamcalendar/texts';
import LicenceSection from './licence/LicenceSection.svelte';
import TeamTabsSection from './teamtabs/TeamTabsSection.svelte';
import { texts as teamTabTexts } from './teamtabs/texts';
import PresenceSection from './presence/PresenceSection.svelte';
import { texts as presenceTexts } from './presence/texts';
import { presenceSearchActions } from './presence/searchActions';
import { parseConfig as parsePresenceConfig } from './presence/presence';
import { serverDe } from './serverDe';
import { lazyMountable, mountable } from './lib/mountable';

export const plugin: WaylightPlugin = {
  apiVersion: PLUGIN_API_VERSION,
  name: 'ee',
  widgets: [
    {
      kind: 'announcements',
      feature: 'announcements',
      category: 'everyday',
      title: announcementTexts.title ?? { en: 'Announcements' },
      description: announcementTexts.description ?? { en: '' },
      keywords: ['announcement', 'notice', 'news', 'ankündigung', 'hinweis', 'aushang'],
      defaultSize: { w: 4, h: 5 },
      icon: 'rss',
      view: lazyMountable(() => import('./announcements/AnnouncementsWidget.svelte')),
    },
    {
      kind: 'absences',
      feature: 'absences',
      category: 'everyday',
      title: absenceTexts.title ?? { en: 'Absences' },
      description: absenceTexts.description ?? { en: '' },
      keywords: ['absence', 'holiday', 'vacation', 'sick', 'abwesenheit', 'urlaub', 'krank', 'team calendar'],
      defaultSize: { w: 4, h: 5 },
      icon: 'calendar',
      view: lazyMountable(() => import('./absences/AbsencesWidget.svelte')),
    },
    {
      kind: 'presence',
      feature: 'presence',
      category: 'everyday',
      title: presenceTexts.title ?? { en: 'Presence' },
      description: presenceTexts.description ?? { en: '' },
      keywords: ['presence', 'office', 'home office', 'who is in', 'anwesenheit', 'büro', 'homeoffice'],
      defaultSize: { w: 4, h: 5 },
      icon: 'activity',
      view: lazyMountable(() => import('./presence/PresenceWidget.svelte')),
      // Small and strip tiles show the group's counts with the own status as chips (#632).
      compact: (config) => {
        const { group } = parsePresenceConfig(config);
        const params: Record<string, string> = group ? { group } : {};
        return { module: 'presence', params };
      },
      compactExample: presenceTexts.compact_example ?? { en: '' },
      settings: lazyMountable<{
        config?: Record<string, unknown>;
        onsave: (config: Record<string, unknown>) => void;
      }>(() => import('./presence/PresenceWidgetSettings.svelte')),
    },
  ],
  settings: [
    {
      id: 'licence',
      area: 'system',
      label: {
        en: 'Features for organisations',
        de: 'Funktionen für Organisationen',
      },
      description: { en: '', de: '' },
      view: mountable(LicenceSection),
    },
    {
      id: 'team-tabs',
      area: 'dashboard',
      feature: 'team-tabs',
      adminOnly: true,
      label: teamTabTexts.label ?? { en: 'Team tabs' },
      description: teamTabTexts.description ?? { en: '' },
      view: mountable(TeamTabsSection),
    },
    {
      id: 'branding',
      area: 'appearance',
      feature: 'branding',
      adminOnly: true,
      label: brandingTexts.label ?? { en: 'Branding' },
      description: brandingTexts.description ?? { en: '' },
      view: lazyMountable(() => import('./branding/BrandingSection.svelte')),
    },
    {
      id: 'connections',
      area: 'connections',
      feature: 'connected-accounts',
      label: connectionTexts.label ?? { en: 'Connected accounts' },
      description: connectionTexts.description ?? { en: '' },
      view: lazyMountable(() => import('./connections/ConnectionsSection.svelte')),
    },
    {
      id: 'team-calendar',
      area: 'dashboard',
      feature: 'absences',
      label: teamCalendarTexts.label ?? { en: 'Team calendar' },
      description: teamCalendarTexts.description ?? { en: '' },
      view: lazyMountable(() => import('./teamcalendar/TeamCalendarSection.svelte')),
    },
    {
      id: 'presence',
      area: 'people',
      feature: 'presence',
      adminOnly: true,
      label: presenceTexts.section ?? { en: 'Presence' },
      description: presenceTexts.section_hint ?? { en: '' },
      view: mountable(PresenceSection),
    },
  ],
  searchActions: [...presenceSearchActions()],
  // The instance name in the first-start flow (#671); only with branding licensed.
  onboardingSteps: [
    {
      id: 'instance-name',
      feature: 'branding',
      ownerOnly: true,
      title: brandingTexts.step_title ?? { en: 'Name of this Waylight' },
      view: lazyMountable<{ setSubmit: (save: () => Promise<void>) => void }>(
        () => import('./branding/InstanceNameStep.svelte'),
      ),
    },
  ],
  // German texts of the ee/ backend messages; the server sends English.
  serverMessages: { de: serverDe },
};
