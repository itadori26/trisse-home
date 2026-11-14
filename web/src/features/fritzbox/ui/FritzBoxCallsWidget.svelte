<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { type CallList, callIcons, callParty, callsOptions } from '../domain/fritzbox';
  import type { FritzBoxUseCases } from '../usecase/fritzboxUseCases';

  /**
   * The FRITZ!Box call list (#168). The server fetches it only for this
   * widget and reuses it for two minutes, so polling every minute is cheap.
   */
  interface Props {
    fritzbox: FritzBoxUseCases;
    config: Record<string, unknown>;
  }
  let { fritzbox, config }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 60_000;
  const options = $derived(callsOptions(config));
  let list = $state<CallList | null>(null);
  let error = $state('');
  let unconfigured = $state(false);

  $effect(() => {
    const o = options;
    const load = () =>
      fritzbox
        .calls(o)
        .then((l) => ((list = l), (error = ''), (unconfigured = false)))
        .catch((err) => {
          unconfigured = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
          throw err;
        });
    return poll(load, POLL_MS);
  });

  const when = (d: Date) =>
    d.toDateString() === new Date().toDateString()
      ? d.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' })
      : d.toLocaleDateString(intlLocale(), { day: '2-digit', month: '2-digit' });
</script>

{#if unconfigured}
  <p class="muted">{t('fritzbox.connect_your_fritz_box_under_settings')}</p>
{:else if error && !list}
  <p class="error">{error}</p>
{:else if list}
  {#if list.calls.length}
    <ul class="calls">
      {#each list.calls as c, i (i)}
        <li
          class:missed={c.type === 'missed'}
          title="{options.showNumbers ? c.number : ''}{c.durationMinutes
            ? ` · ${t('time.minutes', { n: c.durationMinutes })}`
            : ''}"
        >
          <span class="icon" aria-label={t(`fritzbox.call.${c.type}`)}>{callIcons[c.type]}</span>
          <span class="party">{callParty(c, options.showNumbers)}</span>
          <span class="muted small">{when(c.at)}</span>
        </li>
      {/each}
    </ul>
  {:else if !list.hint}
    <p class="muted">{t('fritzbox.no_calls')}</p>
  {/if}
  {#if list.hint}<p class="muted small">{list.hint}</p>{/if}
  {#if list.error}<p class="error small">{list.error}</p>{/if}
{:else}
  <p class="muted">{t('fritzbox.loading')}</p>
{/if}

<style>
  .calls {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .calls li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 1px 0;
  }
  .party {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .missed {
    color: var(--color-danger);
  }
  .icon {
    width: 1.2em;
    text-align: center;
  }
  .small {
    font-size: 0.8em;
  }
</style>
