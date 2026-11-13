<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import type { CalendarEvent } from '../domain/calendar';

  /**
   * Details of one event (#163): time, location, the description as plain
   * text (never HTML) and its link if the event has an http(s) URL.
   */
  interface Props {
    event: CalendarEvent;
    onclose: () => void;
  }
  let { event, onclose }: Props = $props();

  let box = $state<HTMLElement>();
  $effect(() => box?.focus());

  const when = $derived.by(() => {
    const date = (d: Date) =>
      d.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'long' });
    const time = (d: Date) => d.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' });
    if (event.allDay) {
      const last = new Date(event.end.getFullYear(), event.end.getMonth(), event.end.getDate() - 1);
      return last.getTime() > event.start.getTime()
        ? `${date(event.start)} – ${date(last)}`
        : `${date(event.start)} · ${t('calendar.all_day_short')}`;
    }
    return `${date(event.start)} · ${time(event.start)}–${time(event.end)}`;
  });
</script>

<div
  class="details"
  role="dialog"
  aria-label={t('calendar.event_details')}
  tabindex="-1"
  bind:this={box}
  onkeydown={(e) => e.key === 'Escape' && (e.stopPropagation(), onclose())}
>
  <div class="top">
    <strong>{event.summary}</strong>
    <button class="btn btn-icon" type="button" aria-label={t('calendar.close')} onclick={onclose}>✕</button>
  </div>
  <p class="muted">{when}</p>
  {#if event.location}<p>{event.location}</p>{/if}
  {#if event.description}<p class="description">{event.description}</p>{/if}
  {#if event.url}
    <p><a href={event.url} target="_blank" rel="noopener noreferrer">{t('calendar.open_link')}</a></p>
  {/if}
  <p class="muted small">{event.calendar}</p>
</div>

<style>
  .details {
    position: absolute;
    inset: auto 0 0 0;
    z-index: 5;
    max-height: 80%;
    overflow: auto;
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius);
    background: var(--color-surface);
    box-shadow: var(--shadow);
    font-size: 0.9em;
  }
  .top {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-2);
  }
  p {
    margin: var(--space-1) 0 0;
  }
  .description {
    white-space: pre-line;
  }
  .small {
    font-size: 0.8em;
  }
</style>
