<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { describe, type Forecast, parseWeatherConfig } from '../domain/weather';
  import type { WeatherUseCases } from '../usecase/weatherUseCases';

  /** Props of the weather widget. */
  interface Props {
    weather: WeatherUseCases;
    config: Record<string, unknown>;
  }
  let { weather, config }: Props = $props();

  /** The backend refreshes every 15 minutes; polling more often is pointless. */
  const POLL_MS = 10 * 60 * 1000;
  const cfg = $derived(parseWeatherConfig(config));
  let forecast = $state<Forecast | null>(null);
  let error = $state('');

  $effect(() => {
    const c = cfg;
    if (!c) return;
    const load = () =>
      weather
        .forecast(c)
        .then((f) => ((forecast = f), (error = '')))
        .catch((err) => (error = errorMessage(err)));
    load();
    const timer = setInterval(load, POLL_MS);
    return () => clearInterval(timer);
  });

  const weekday = (d: Date) => d.toLocaleDateString(undefined, { weekday: 'short' });
  const deg = (v: number) => `${Math.round(v)}°`;
</script>

{#if !cfg}
  <p class="muted">Choose a location in the widget settings.</p>
{:else if error && !forecast}
  <p class="error">{error}</p>
{:else if forecast}
  {@const now = describe(forecast.code, forecast.isDay)}
  <div class="weather">
    <div class="now">
      <span class="icon" aria-hidden="true">{now.icon}</span>
      <div>
        <div class="temp">{deg(forecast.temperature)}</div>
        <div class="muted">{now.label}{cfg.label ? ` · ${cfg.label}` : ''}</div>
        <div class="muted small details">
          Feels {deg(forecast.apparentTemperature)} · {Math.round(forecast.humidity)}% · {Math.round(
            forecast.windSpeed,
          )} km/h
        </div>
      </div>
    </div>
    {#if forecast.days.length}
      <ul class="days">
        {#each forecast.days as d (d.date.getTime())}
          {@const c = describe(d.code)}
          <li title={c.label}>
            <span class="muted">{weekday(d.date)}</span>
            <span class="day-icon" aria-hidden="true">{c.icon}</span>
            <span>{deg(d.max)} <span class="muted">{deg(d.min)}</span></span>
            {#if d.precipitationProbability > 0}<span class="rain small"
                >{Math.round(d.precipitationProbability)}%</span
              >{/if}
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{:else}
  <p class="muted">Loading…</p>
{/if}

<style>
  .weather {
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: space-around;
    gap: var(--space-3);
  }
  .now {
    display: flex;
    gap: 0.6em;
    align-items: center;
    justify-content: center;
  }
  .icon {
    font-size: clamp(2rem, 20cqmin, 8rem);
    line-height: 1;
  }
  .temp {
    font-size: clamp(1.6rem, 15cqmin, 6rem);
    font-weight: 600;
    line-height: 1;
  }
  .small {
    font-size: 0.85em;
  }
  .days {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(3.5em, 1fr));
    gap: var(--space-2);
    text-align: center;
  }
  .days li {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .day-icon {
    font-size: clamp(1.2rem, 8cqmin, 3rem);
  }
  .rain {
    color: var(--color-accent);
  }
  /* Wide widgets: forecast next to the current conditions. */
  @container widget (min-aspect-ratio: 2.2 / 1) and (min-width: 520px) {
    .weather {
      flex-direction: row;
      align-items: center;
    }
    .days {
      flex: 1;
    }
  }
  /* Low widgets drop the forecast, very low ones the details as well. */
  @container widget (max-height: 150px) and (max-width: 519px) {
    .days {
      display: none;
    }
  }
  @container widget (max-height: 90px) {
    .details {
      display: none;
    }
  }
</style>
