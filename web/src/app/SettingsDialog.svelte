<script lang="ts">
  import { TimelineSettings } from '../features/timeline';
  import { BriefingSettings } from '../features/briefing';
  import { AboutSettings, ForecastSettings } from '../features/system';
  import { type MessageKey, t } from '../shared/i18n';
  import { AdGuardSettings } from '../features/adguard';
  import { MediaSettings } from '../features/media';
  import { SmartHomeSettings } from '../features/smarthome';
  import { ProxmoxSettings } from '../features/proxmox';
  import { SpeedtestSettings } from '../features/speedtest';
  import { ImageUpdateSettings } from '../features/docker';
  import { FritzBoxSettings } from '../features/fritzbox';
  import { NextcloudSettings } from '../features/nextcloud';
  import { ShoppingSettings, suggestListUrl } from '../features/shopping';
  import { ParcelsSettings } from '../features/parcels';
  import { TransferSettings } from '../features/transfer';
  import { WasteSettings } from '../features/waste';
  import { WireGuardSettings } from '../features/wireguard';
  import { type Appearance, AppearanceSettings } from '../features/appearance';
  import {
    ApiTokensSettings,
    ChangePasswordForm,
    PasskeySettings,
    GuestsSettings,
    SecurityLog,
    SessionsList,
    TwoFactorSettings,
  } from '../features/auth';
  import { BackupSettings } from '../features/backup';
  import { BackupSourcesManager } from '../features/backupmonitor';
  import { NotificationSettings } from '../features/notifications';
  import { BookmarkImport, BookmarkLinkSettings, FoldersManager } from '../features/bookmarks';
  import { CalendarSettings } from '../features/calendar';
  import { FeedsManager } from '../features/rss';
  import { SearchSettings } from '../features/search';
  import { ServicesManager } from '../features/services';
  import Modal from '../shared/ui/Modal.svelte';
  import type { Container } from './container';

  /** Props of the global settings dialog. */
  interface Props {
    container: Container;
    modules: Set<string>;
    appearance: Appearance;
    onappearance: (a: Appearance) => void;
    /** Restarts the guided tour; omitted when the onboarding module is disabled. */
    onrestarttour?: () => void;
    onclose: () => void;
  }
  let { container, modules, appearance, onappearance, onrestarttour, onclose }: Props = $props();

  type Section = { id: string; label: MessageKey; module?: string };
  const sections: Section[] = [
    { id: 'appearance', label: 'settings.section.appearance', module: 'appearance' },
    { id: 'search', label: 'settings.section.search', module: 'search' },
    { id: 'services', label: 'settings.section.services', module: 'services' },
    { id: 'feeds', label: 'settings.section.feeds', module: 'rss' },
    { id: 'folders', label: 'settings.section.bookmark_folders', module: 'bookmarks' },
    { id: 'calendar', label: 'settings.section.calendar', module: 'calendar' },
    { id: 'adguard', label: 'settings.section.adguard', module: 'adguard' },
    { id: 'media', label: 'settings.section.media', module: 'media' },
    { id: 'smarthome', label: 'settings.section.smarthome', module: 'smarthome' },
    { id: 'proxmox', label: 'settings.section.proxmox', module: 'proxmox' },
    { id: 'speedtest', label: 'settings.section.speedtest', module: 'speedtest' },
    { id: 'docker-updates', label: 'settings.section.docker_updates', module: 'docker' },
    { id: 'nextcloud', label: 'settings.section.nextcloud', module: 'nextcloud' },
    { id: 'fritzbox', label: 'settings.section.fritzbox', module: 'fritzbox' },
    { id: 'wireguard', label: 'settings.section.wireguard', module: 'wireguard' },
    { id: 'waste', label: 'settings.section.waste', module: 'waste' },
    { id: 'shopping', label: 'settings.section.shopping', module: 'shopping' },
    { id: 'parcels', label: 'settings.section.parcels', module: 'parcels' },
    { id: 'transfer', label: 'settings.section.transfer', module: 'transfer' },
    { id: 'notifications', label: 'settings.section.notifications', module: 'notifications' },
    { id: 'timeline', label: 'settings.section.timeline', module: 'timeline' },
    { id: 'briefing', label: 'settings.section.briefing', module: 'briefing' },
    { id: 'backups', label: 'settings.section.backups', module: 'backup' },
    { id: 'backupmonitor', label: 'settings.section.backupmonitor', module: 'backupmonitor' },
    { id: 'account', label: 'settings.section.account' },
    { id: 'security-log', label: 'settings.section.security_log' },
    { id: 'guests', label: 'settings.section.guests' },
    { id: 'tokens', label: 'settings.section.tokens' },
    { id: 'forecast', label: 'settings.section.forecast', module: 'system' },
    { id: 'about', label: 'settings.section.about', module: 'system' },
  ];
  /** Guests only manage their own account (ADR 0012). */
  const guestSections = new Set(['account']);
  const visible = $derived(
    sections.filter(
      (s) => (!s.module || modules.has(s.module)) && (!container.auth.guest || guestSections.has(s.id)),
    ),
  );
  let guestTabs = $state<{ id: string; name: string }[]>([]);
  $effect(() => {
    if (!container.auth.guest) container.dashboard.load().then((tabs) => (guestTabs = tabs));
  });
  let active = $state('');
  const current = $derived(active || visible[0]?.id || 'account');
</script>

<Modal title={t('app.settings')} {onclose}>
  <nav class="sections">
    {#each visible as s (s.id)}
      <button class="btn" class:btn-primary={current === s.id} type="button" onclick={() => (active = s.id)}
        >{t(s.label)}</button
      >
    {/each}
  </nav>
  {#if current === 'appearance'}
    <AppearanceSettings appearance={container.appearance} current={appearance} onchange={onappearance} />
  {:else if current === 'search'}
    <SearchSettings search={container.search} />
  {:else if current === 'services'}
    <ServicesManager services={container.services} />
  {:else if current === 'feeds'}
    <FeedsManager rss={container.rss} />
  {:else if current === 'folders'}
    <FoldersManager bookmarks={container.bookmarks} />
    <BookmarkLinkSettings bookmarks={container.bookmarks} />
  {:else if current === 'calendar'}
    <CalendarSettings calendar={container.calendar} />
  {:else if current === 'transfer'}
    <TransferSettings transfer={container.transfer} />
    {#if modules.has('bookmarks')}<BookmarkImport bookmarks={container.bookmarks} />{/if}
  {:else if current === 'security-log'}
    <SecurityLog auth={container.auth} />
  {:else if current === 'tokens'}
    <ApiTokensSettings auth={container.auth} {modules} />
  {:else if current === 'guests'}
    <GuestsSettings auth={container.auth} tabs={guestTabs} />
  {:else if current === 'timeline'}
    <TimelineSettings timeline={container.timeline} />
  {:else if current === 'briefing'}
    <BriefingSettings briefing={container.briefing} canPush={modules.has('notifications')} />
  {:else if current === 'forecast'}
    <ForecastSettings system={container.system} />
  {:else if current === 'about'}
    <AboutSettings system={container.system} />
  {:else if current === 'parcels'}
    <ParcelsSettings parcels={container.parcels} />
  {:else if current === 'shopping'}
    <ShoppingSettings
      shopping={container.shopping}
      suggest={async () => {
        const c = await container.calendar.settings();
        return c.configured
          ? { listUrl: suggestListUrl(c.calendarUrls[0] ?? ''), username: c.username }
          : null;
      }}
    />
  {:else if current === 'waste'}
    <WasteSettings waste={container.waste} />
  {:else if current === 'wireguard'}
    <WireGuardSettings wireguard={container.wireguard} />
  {:else if current === 'fritzbox'}
    <FritzBoxSettings fritzbox={container.fritzbox} />
  {:else if current === 'nextcloud'}
    <NextcloudSettings nextcloud={container.nextcloud} />
  {:else if current === 'docker-updates'}
    <ImageUpdateSettings docker={container.docker} />
  {:else if current === 'speedtest'}
    <SpeedtestSettings speedtest={container.speedtest} />
  {:else if current === 'adguard'}
    <AdGuardSettings adguard={container.adguard} />
  {:else if current === 'media'}
    <MediaSettings media={container.media} />
  {:else if current === 'smarthome'}
    <SmartHomeSettings smarthome={container.smarthome} />
  {:else if current === 'proxmox'}
    <ProxmoxSettings proxmox={container.proxmox} />
  {:else if current === 'notifications'}
    <NotificationSettings notifications={container.notifications} />
  {:else if current === 'backupmonitor'}
    <BackupSourcesManager backupMonitor={container.backupMonitor} />
  {:else if current === 'backups'}
    <BackupSettings backups={container.backup} />
  {:else}
    {#if onrestarttour}
      <h3>{t('app.guided_tour')}</h3>
      <p class="muted">{t('app.shows_the_most_important_features_of')}</p>
      <button class="btn" type="button" onclick={onrestarttour}>{t('app.restart_the_tour')}</button>
    {/if}
    <h3>{t('app.two_factor_authentication')}</h3>
    <TwoFactorSettings auth={container.auth} />
    <PasskeySettings auth={container.auth} />
    <h3>{t('app.active_sessions')}</h3>
    <SessionsList auth={container.auth} />
    <h3>{t('app.change_password')}</h3>
    <ChangePasswordForm auth={container.auth} />
  {/if}
</Modal>

<style>
  .sections {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin-bottom: var(--space-4);
  }
</style>
