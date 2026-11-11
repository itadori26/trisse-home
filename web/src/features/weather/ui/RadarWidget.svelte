<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseWeatherConfig, radarKey } from '../domain/weather';
  import type { WeatherUseCases } from '../usecase/weatherUseCases';

  /** Props of the rain radar widget. */
  interface Props {
    weather: WeatherUseCases;
    config: Record<string, unknown>;
  }
  let { weather, config }: Props = $props();

  const cfg = $derived(parseWeatherConfig(config));
  let key = $state(radarKey(new Date()));
  let failed = $state(false);

  $effect(() => {
    const timer = setInterval(() => {
      key = radarKey(new Date());
      failed = false;
    }, 60_000);
    return () => clearInterval(timer);
  });
</script>

{#if !cfg}
  <p class="muted">{t('weather.choose_a_location_in_the_widget')}</p>
{:else if failed}
  <p class="muted">{t('weather.the_rain_radar_is_not_available')}</p>
{:else}
  <figure class="radar">
    <img
      src={weather.radarUrl(cfg, key)}
      alt={t('weather.rain_radar_around', { value: cfg.label || t('weather.the_location') })}
      onerror={() => (failed = true)}
    />
    <span class="marker" aria-hidden="true"></span>
    <figcaption>
      {t('weather.deutscher_wetterdienst', { value: cfg.label ? `${cfg.label} · ` : '' })}
    </figcaption>
  </figure>
{/if}

<style>
  .radar {
    position: relative;
    margin: 0;
    height: 100%;
    display: flex;
    flex-direction: column;
    background: #1f2937;
    border-radius: var(--radius-small);
    overflow: hidden;
  }
  img {
    flex: 1;
    min-height: 0;
    width: 100%;
    object-fit: cover;
  }
  .marker {
    position: absolute;
    left: 50%;
    top: calc(50% - 0.7em);
    width: 10px;
    height: 10px;
    margin: -5px 0 0 -5px;
    border-radius: 50%;
    background: #ef4444;
    box-shadow: 0 0 0 2px #fff;
  }
  figcaption {
    font-size: 0.7em;
    color: #e5e7eb;
    padding: 2px 6px;
  }
</style>
