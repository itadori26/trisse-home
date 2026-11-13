<script lang="ts">
  import type { TodosUseCases } from '../features/todos/usecase/todosUseCases';
  import { describe, type WeatherConfig } from '../features/weather/domain/weather';
  import type { WeatherUseCases } from '../features/weather/usecase/weatherUseCases';
  import { intlLocale, t } from '../shared/i18n';
  import { usePolling } from '../shared/refresh';

  /** Props of the centre of the top bar (#189): date, weather and due to-dos. */
  interface Props {
    /** Null when the module is disabled. */
    weather: WeatherUseCases | null;
    /** Location of the first weather widget, or null. */
    location: WeatherConfig | null;
    todos: TodosUseCases | null;
  }
  let { weather, location, todos }: Props = $props();
  const poll = usePolling();

  let now = $state(new Date());
  let temperature = $state<{ value: number; icon: string } | null>(null);
  let due = $state<number | null>(null);

  $effect(() => {
    const id = setInterval(() => (now = new Date()), 60_000);
    return () => clearInterval(id);
  });

  // The weather use case shares its forecast with the weather widget (#167).
  $effect(() => {
    const loc = location;
    temperature = null;
    if (!weather || !loc) return;
    let cancelled = false;
    weather
      .forecast(loc)
      .then((f) => {
        if (!cancelled)
          temperature = { value: Math.round(f.temperature), icon: describe(f.code, f.isDay).icon };
      })
      .catch(() => {});
    return () => (cancelled = true);
  });

  $effect(() => {
    if (!todos) return;
    const uc = todos;
    return poll(async () => {
      const today = new Date();
      const iso = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
      const list = await uc.list(false);
      due = list.filter((x) => x.due !== '' && x.due <= iso).length;
    }, 5 * 60_000);
  });

  const date = $derived(
    now.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'long' }),
  );
</script>

<p class="info">
  <span>{date}</span>
  {#if temperature}
    <span class="sep" aria-hidden="true">·</span>
    <span>{temperature.icon} {t('app.greeting_temperature', { t: temperature.value })}</span>
  {/if}
  {#if due}
    <span class="sep" aria-hidden="true">·</span>
    <span>{t(due === 1 ? 'app.todo_due_one' : 'app.todos_due', { n: due })}</span>
  {/if}
</p>

<style>
  .info {
    margin: 0;
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0 var(--space-2);
    color: var(--color-text-muted);
    white-space: nowrap;
  }
</style>
