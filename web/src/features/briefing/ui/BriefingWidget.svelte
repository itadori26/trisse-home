<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { requestWidgetFocus } from '../../../shared/ui/widgetFocus';
  import { type Briefing, itemDetail, itemTime, sectionTitle, SOURCE_WIDGETS } from '../domain/briefing';
  import type { BriefingUseCases } from '../usecase/briefingUseCases';

  /** Props of the daily briefing widget (#52). */
  interface Props {
    briefing: BriefingUseCases;
  }
  let { briefing }: Props = $props();

  const poll = usePolling();

  let data = $state<Briefing | null>(null);
  let error = $state('');

  async function load() {
    try {
      data = await briefing.get();
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err; // lets the scheduler back off
    }
  }

  // The briefing changes with the time of day; refresh every 5 minutes
  // through the shared scheduler, which pauses hidden pages (#117).
  $effect(() => poll(load, 5 * 60_000));
</script>

<div class="briefing" aria-busy={data === null && !error}>
  {#if error}<p class="error">{error}</p>{/if}
  {#if data}
    {#each data.sections as s (s.kind)}
      <section class="section" class:alert={s.items.some((i) => i.severity === 'critical')}>
        <h4>
          <button
            type="button"
            class="link"
            title={t('briefing.open_source')}
            onclick={() => requestWidgetFocus(SOURCE_WIDGETS[s.kind])}>{sectionTitle(s.kind)}</button
          >
        </h4>
        <ul>
          {#each s.items as item, i (i)}
            {@const time = itemTime(s.kind, item, intlLocale())}
            {@const detail = itemDetail(s.kind, item)}
            <li class={item.severity}>
              {#if time}<span class="time">{time}</span>{/if}
              <span class="text">{item.text}</span>
              {#if detail}<span class="detail">{detail}</span>{/if}
            </li>
          {/each}
          {#if s.more > 0}<li class="detail">{t('briefing.more', { n: s.more })}</li>{/if}
        </ul>
      </section>
    {:else}
      <p class="muted">{t('briefing.nothing')}</p>
    {/each}
    {#if data.unavailable.length > 0}
      <p class="muted small">
        {t('briefing.unavailable', { sections: data.unavailable.map(sectionTitle).join(', ') })}
      </p>
    {/if}
  {/if}
</div>

<style>
  .briefing {
    display: grid;
    gap: var(--space-2);
  }
  .section h4 {
    margin: 0 0 2px;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted);
  }
  .section.alert h4 {
    color: var(--color-danger);
  }
  .link {
    border: none;
    background: none;
    padding: 0;
    font: inherit;
    color: inherit;
    cursor: pointer;
    text-transform: inherit;
  }
  .link:hover {
    text-decoration: underline;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    gap: var(--space-2);
    align-items: baseline;
    min-width: 0;
  }
  .text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .time {
    font-variant-numeric: tabular-nums;
    color: var(--color-text-muted);
  }
  .detail {
    font-size: 0.85em;
    color: var(--color-text-muted);
    white-space: nowrap;
  }
  li.warning .detail,
  li.warning .text {
    color: var(--color-warning, #b26a00);
  }
  li.critical .text {
    color: var(--color-danger);
    font-weight: 600;
  }
  .small {
    font-size: 0.8rem;
    margin: 0;
  }
</style>
