<script lang="ts">
  import { dayPart, greetingKey } from '../features/appearance/domain/greeting';
  import type { WeatherConfig } from '../features/weather/domain/weather';
  import type { WeatherUseCases } from '../features/weather/usecase/weatherUseCases';
  import { intlLocale, t } from '../shared/i18n';

  /** Props of the greeting line below the top bar (#167). */
  interface Props {
    name: string;
    /** Weather use cases, or null when the weather module is disabled. */
    weather: WeatherUseCases | null;
    /** Location of the first weather widget, or null. */
    location: WeatherConfig | null;
  }
  let { name, weather, location }: Props = $props();

  let now = $state(new Date());
  let temperature = $state<number | null>(null);

  // The greeting and the date only change every few minutes.
  $effect(() => {
    const id = setInterval(() => (now = new Date()), 60_000);
    return () => clearInterval(id);
  });

  $effect(() => {
    const loc = location;
    temperature = null;
    if (!weather || !loc) return;
    let cancelled = false;
    weather
      .forecast(loc)
      .then((f) => !cancelled && (temperature = Math.round(f.temperature)))
      .catch(() => {}); // the line simply shows no temperature
    return () => (cancelled = true);
  });

  const hello = $derived(t(greetingKey(dayPart(now), name !== ''), { name }));
  const date = $derived(
    now.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'long' }),
  );
</script>

<p class="greeting">
  <span class="hello">{hello}</span>
  <span class="sep" aria-hidden="true">·</span>
  <span>{date}</span>
  {#if temperature !== null}
    <span class="sep" aria-hidden="true">·</span>
    <span>{t('app.greeting_temperature', { t: temperature })}</span>
  {/if}
</p>

<style>
  .greeting {
    margin: 0 auto;
    max-width: 1600px;
    padding: 0 var(--space-4) var(--space-2);
    color: var(--color-text-muted);
    font-size: 0.9rem;
    display: flex;
    flex-wrap: wrap;
    gap: 0 var(--space-2);
  }
  .hello {
    color: var(--color-text);
    font-weight: 500;
  }
</style>
