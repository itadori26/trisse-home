<script lang="ts">
  import { t } from '../../shared/i18n';
  import { inNightWindow, type KioskOptions } from '../../features/auth/domain/auth';

  /**
   * Kiosk display behaviour (#50): no controls, optional tab rotation and
   * night dimming, full screen, screen kept awake, cursor hidden when idle.
   */
  interface Props {
    loadOptions: () => Promise<KioskOptions>;
    tabCount: number;
    activeIndex: number;
    selectTabAt: (index: number) => void;
  }
  let { loadOptions, tabCount, activeIndex, selectTabAt }: Props = $props();

  let options = $state<KioskOptions>({ rotateSeconds: 0, nightFrom: '', nightTo: '' });
  let now = $state(new Date());
  let fullscreen = $state(typeof document !== 'undefined' && !!document.fullscreenElement);
  const night = $derived(inNightWindow(now, options.nightFrom, options.nightTo));

  $effect(() => {
    loadOptions()
      .then((o) => (options = o))
      .catch(() => undefined);
  });

  // The page shows the display mode; html classes hide the controls (CSS below).
  $effect(() => {
    document.documentElement.classList.add('kiosk');
    return () => document.documentElement.classList.remove('kiosk', 'kiosk-idle');
  });

  $effect(() => {
    const id = setInterval(() => (now = new Date()), 30_000);
    return () => clearInterval(id);
  });

  // Tab rotation.
  $effect(() => {
    const seconds = options.rotateSeconds;
    const count = tabCount;
    if (!seconds || count < 2) return;
    const id = setTimeout(() => selectTabAt((activeIndex + 1) % count), seconds * 1000);
    return () => clearTimeout(id);
  });

  // Keep the screen on where the Wake Lock API exists; it is released when
  // the page is hidden and asked for again when it comes back.
  $effect(() => {
    let lock: { release: () => Promise<void> } | null = null;
    const nav = navigator as Navigator & {
      wakeLock?: { request: (t: 'screen') => Promise<{ release: () => Promise<void> }> };
    };
    const acquire = () => {
      if (document.visibilityState === 'visible' && nav.wakeLock) {
        nav.wakeLock
          .request('screen')
          .then((l) => (lock = l))
          .catch(() => undefined);
      }
    };
    acquire();
    document.addEventListener('visibilitychange', acquire);
    return () => {
      document.removeEventListener('visibilitychange', acquire);
      void lock?.release().catch(() => undefined);
    };
  });

  // Hide the cursor after three seconds without movement.
  $effect(() => {
    let timer: ReturnType<typeof setTimeout>;
    const wake = () => {
      document.documentElement.classList.remove('kiosk-idle');
      clearTimeout(timer);
      timer = setTimeout(() => document.documentElement.classList.add('kiosk-idle'), 3000);
    };
    wake();
    window.addEventListener('pointermove', wake);
    return () => {
      clearTimeout(timer);
      window.removeEventListener('pointermove', wake);
    };
  });

  $effect(() => {
    const update = () => (fullscreen = !!document.fullscreenElement);
    document.addEventListener('fullscreenchange', update);
    return () => document.removeEventListener('fullscreenchange', update);
  });

  // Full screen needs a user gesture, so the display offers a button once.
  const enterFullscreen = () => void document.documentElement.requestFullscreen?.().catch(() => undefined);
</script>

{#if !fullscreen && typeof document !== 'undefined' && document.fullscreenEnabled}
  <button class="btn fullscreen" type="button" onclick={enterFullscreen}>⛶ {t('kiosk.fullscreen')}</button>
{/if}
{#if night}<div class="night" aria-hidden="true"></div>{/if}

<style>
  .fullscreen {
    position: fixed;
    right: var(--space-4);
    bottom: var(--space-4);
    z-index: 50;
  }
  .night {
    position: fixed;
    inset: 0;
    background: rgb(0 0 0 / 70%);
    pointer-events: none;
    z-index: 40;
  }
  /* No controls on a display: launcher, account menu, search and status. */
  :global(html.kiosk .topbar .right),
  :global(html.kiosk .hero .pill),
  :global(html.kiosk [data-action]) {
    display: none !important;
  }
  :global(html.kiosk-idle),
  :global(html.kiosk-idle *) {
    cursor: none !important;
  }
</style>
