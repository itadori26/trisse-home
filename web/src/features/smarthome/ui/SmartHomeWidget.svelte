<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { changedAgo, type Entity, isOn, parseWidgetConfig, stateText } from '../domain/smarthome';
  import type { SmarthomeUseCases } from '../usecase/smarthomeUseCases';

  /** Props of the smart home widget (#49). */
  interface Props {
    smarthome: SmarthomeUseCases;
    config: Record<string, unknown>;
  }
  let { smarthome, config }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 30_000;
  const ids = $derived(parseWidgetConfig(config).entities);
  let entities = $state<Entity[] | null>(null);
  let error = $state('');
  let unconfigured = $state(false);
  let busy = $state('');
  let now = $state(new Date());

  const load = () =>
    smarthome
      .states(ids)
      .then((list) => ((entities = list), (error = ''), (unconfigured = false), (now = new Date())))
      .catch((err) => {
        unconfigured = err instanceof ApiError && err.status === 404;
        error = errorMessage(err);
        throw err;
      });

  $effect(() => {
    void ids;
    return poll(load, POLL_MS);
  });

  async function act(e: Entity) {
    const question = e.action === 'press' ? 'smarthome.confirm_press' : 'smarthome.confirm_toggle';
    if (!confirm(t(question, { name: e.name }))) return;
    busy = e.id;
    try {
      await smarthome.act(e.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = '';
    }
  }
</script>

{#if unconfigured}
  <p class="muted">{t('smarthome.connect_under_settings')}</p>
{:else if ids.length === 0}
  <p class="muted">{t('smarthome.choose_entities')}</p>
{:else if error && !entities}
  <p class="error">{error}</p>
{:else if entities}
  <ul class="tiles">
    {#each entities as e (e.id)}
      <li class:on={isOn(e)}>
        <span class="name" title={e.id}>{e.name}</span>
        <strong class="state">{stateText(e)}</strong>
        <span class="muted small">{changedAgo(e.lastChanged, now)}</span>
        {#if e.action}
          <button
            class="btn"
            type="button"
            disabled={busy !== ''}
            aria-pressed={e.action === 'toggle' ? isOn(e) : undefined}
            onclick={() => act(e)}
            >{busy === e.id
              ? '…'
              : e.action === 'press'
                ? t('smarthome.press')
                : isOn(e)
                  ? t('smarthome.turn_off')
                  : t('smarthome.turn_on')}</button
          >
        {/if}
      </li>
    {/each}
  </ul>
  {#if error}<p class="error small">{error}</p>{/if}
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .tiles {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(9rem, 1fr));
    gap: var(--space-2);
  }
  .tiles li {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius, 8px);
  }
  .tiles li.on {
    border-color: var(--color-accent);
  }
  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .state {
    font-size: 1.2em;
  }
  .small {
    font-size: 0.8em;
  }
</style>
