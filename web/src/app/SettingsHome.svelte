<script lang="ts">
  import { t } from '../shared/i18n';
  import Icon from '../shared/ui/icons/Icon.svelte';
  import type { AreaInfo, SettingsArea } from './settingsNav';
  import type { AreaStatus } from './settingsStatus';
  import { attention } from './attention.svelte';

  /** Props of the settings start page (#600). */
  interface Props {
    /** The areas the viewer can open, in display order. */
    areas: readonly AreaInfo[];
    statuses: Partial<Record<SettingsArea, AreaStatus>>;
    /** Shortcuts to the sections changed last; only ones the viewer may open. */
    recent: readonly { id: string; name: string; area: string }[];
    onarea: (id: SettingsArea) => void;
    onsection: (id: string) => void;
  }
  let { areas, statuses, recent, onarea, onsection }: Props = $props();
</script>

<!-- What stopped working or is not set up yet (#592), above everything else. -->
{#if attention.count > 0}
  <button type="button" class="attention" onclick={() => onsection('attention')}>
    <span class="dot" aria-hidden="true"></span>
    {attention.count === 1 ? t('attention.badge_one') : t('attention.badge', { n: attention.count })}
  </button>
{/if}

<ul class="tiles" aria-label={t('settings.areas')}>
  {#each areas as a (a.id)}
    {@const status = statuses[a.id]}
    <li>
      <button type="button" class="tile" data-area={a.id} onclick={() => onarea(a.id)}>
        <span class="icon"><Icon name={a.icon} size={22} /></span>
        <span class="text">
          <span class="label">{t(a.label)}</span>
          <span class="hint">{t(a.hint)}</span>
          {#if status}
            <span class="status {status.tone}">
              {#if status.tone === 'warn' || status.tone === 'error'}<span class="dot" aria-hidden="true"
                ></span>{/if}
              {status.text}
            </span>
          {/if}
        </span>
      </button>
    </li>
  {/each}
</ul>

{#if recent.length > 0}
  <h3 class="recent-title" id="settings-recent">{t('settings.recent')}</h3>
  <ul class="recent" aria-labelledby="settings-recent">
    {#each recent as r (r.id)}
      <li>
        <button type="button" class="btn" onclick={() => onsection(r.id)}>
          {r.name}<span class="muted"> · {r.area}</span>
        </button>
      </li>
    {/each}
  </ul>
{/if}

<style>
  .attention {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    margin-bottom: var(--space-3);
    padding: var(--space-2) var(--space-3);
    font: inherit;
    font-weight: 600;
    text-align: left;
    color: inherit;
    background: color-mix(in srgb, var(--color-warning) 18%, transparent);
    border: 1px solid var(--color-warning);
    border-radius: var(--radius-small);
    cursor: pointer;
  }
  .attention .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-warning);
  }
  .tiles,
  .recent {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .tiles {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
    gap: var(--space-2);
  }
  .tile {
    display: flex;
    gap: var(--space-3);
    align-items: flex-start;
    width: 100%;
    height: 100%;
    padding: var(--space-3);
    text-align: left;
    font: inherit;
    color: inherit;
    background: var(--color-surface-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    cursor: pointer;
  }
  .tile:hover {
    border-color: var(--color-accent);
  }
  .tile:focus-visible {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 2px;
  }
  .icon {
    flex: none;
    color: var(--color-accent-text);
  }
  .text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .label {
    font-weight: 600;
  }
  .hint {
    font-size: 0.85rem;
    color: var(--color-text-muted);
  }
  .status {
    margin-top: var(--space-1);
    font-size: 0.85rem;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .status.ok {
    color: var(--color-success);
  }
  .status.info {
    color: var(--color-text-muted);
  }
  .status.warn {
    color: var(--color-warning);
  }
  .status.error {
    color: var(--color-danger);
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: currentColor;
  }
  .recent-title {
    margin: var(--space-4) 0 var(--space-2);
    font-size: 0.75rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--color-text-muted);
  }
  .recent {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  @media (max-width: 640px) {
    .tiles {
      grid-template-columns: 1fr;
    }
  }
</style>
