<script lang="ts">
  import { t } from '../../../shared/i18n';
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import type { IconName } from '../../../shared/ui/icons/paths';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { groupByCategory, moveFocus, searchWidgetTypes } from '../domain/catalog';
  import type { WidgetType } from '../domain/dashboard';

  /** Props of the widget picker (#104). */
  interface Props {
    types: WidgetType[];
    icon: (kind: string) => IconName | undefined;
    onpick: (type: WidgetType) => void;
    onclose: () => void;
  }
  let { types, icon, onpick, onclose }: Props = $props();

  let query = $state('');
  let search = $state<HTMLInputElement>();
  const matches = $derived(searchWidgetTypes(types, query));
  // Ranked results keep their rank inside each category column.
  const groups = $derived(groupByCategory(matches));
  const tiles: HTMLButtonElement[][] = [];

  $effect(() => {
    search?.focus();
  });

  function focusTile(c: number, r: number) {
    tiles[c]?.[r]?.focus();
  }

  function onTileKey(e: KeyboardEvent, c: number, r: number) {
    if (!['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(e.key)) return;
    e.preventDefault();
    if (e.key === 'ArrowUp' && c === 0 && r === 0) {
      search?.focus();
      return;
    }
    const [nc, nr] = moveFocus(
      groups.map((g) => g.types.length),
      [c, r],
      e.key as Parameters<typeof moveFocus>[2],
    );
    focusTile(nc, nr);
  }

  function onSearchKey(e: KeyboardEvent) {
    if (e.key === 'ArrowDown' && groups.length) {
      e.preventDefault();
      focusTile(0, 0);
    } else if (e.key === 'Enter' && matches[0]) {
      e.preventDefault();
      onpick(matches[0]);
    }
  }
</script>

<Modal title={t('dashboard.add_widget_2')} {onclose} size="wide">
  <div class="picker">
    <input
      class="input search"
      type="search"
      bind:this={search}
      bind:value={query}
      onkeydown={onSearchKey}
      placeholder={t('picker.search')}
      aria-label={t('picker.search')}
      aria-controls="widget-picker-grid"
    />
    <p class="sr-only" role="status" aria-live="polite">{t('picker.count', { count: matches.length })}</p>
    {#if groups.length === 0}
      <p class="muted empty">{t('picker.no_match', { query })}</p>
    {:else}
      <div class="columns" id="widget-picker-grid" role="group" aria-label={t('picker.widgets')}>
        {#each groups as g, c (g.category)}
          <section class="column" aria-labelledby="picker-cat-{g.category}">
            <h3 id="picker-cat-{g.category}">{t(`picker.category.${g.category}`)}</h3>
            <ul>
              {#each g.types as type, r (type.kind)}
                {@const ic = icon(type.kind)}
                <li>
                  <button
                    class="tile"
                    type="button"
                    bind:this={
                      () => tiles[c]?.[r],
                      (el) => {
                        tiles[c] ??= [];
                        if (el) tiles[c][r] = el;
                      }
                    }
                    onclick={() => onpick(type)}
                    onkeydown={(e) => onTileKey(e, c, r)}
                  >
                    {#if ic}<span class="icon"><Icon name={ic} size={28} /></span>{/if}
                    <strong>{type.title}</strong>
                    <span class="desc">{type.description}</span>
                  </button>
                </li>
              {/each}
            </ul>
          </section>
        {/each}
      </div>
    {/if}
  </div>
</Modal>

<style>
  .picker {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    max-height: 70vh;
  }
  .search {
    width: 100%;
  }
  .columns {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--space-4);
    overflow-y: auto;
    padding: 2px;
  }
  h3 {
    margin: 0 0 var(--space-2);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-muted, #888);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-2);
  }
  .tile {
    display: grid;
    grid-template-columns: auto 1fr;
    grid-template-rows: auto auto;
    column-gap: var(--space-2);
    align-items: center;
    width: 100%;
    min-height: 4.5rem;
    padding: var(--space-2);
    text-align: left;
    border: 1px solid var(--color-border, rgba(127, 127, 127, 0.3));
    border-radius: var(--radius-small);
    background: var(--color-surface-2);
    color: inherit;
    font: inherit;
    cursor: pointer;
    transition:
      border-color 0.1s,
      transform 0.1s;
  }
  .tile:hover {
    border-color: var(--color-accent);
  }
  .tile:focus-visible {
    outline: 2px solid var(--color-accent);
    outline-offset: 2px;
  }
  .tile:active {
    transform: scale(0.98);
  }
  .icon {
    grid-row: 1 / span 2;
    display: flex;
    color: var(--color-accent);
  }
  .desc {
    grid-column: 2;
    font-size: 0.8rem;
    color: var(--color-muted, #888);
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
  }
  .empty {
    text-align: center;
    padding: var(--space-6) 0;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
  }
  @media (max-width: 640px) {
    .picker {
      max-height: none;
    }
    .columns {
      grid-template-columns: 1fr;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .tile {
      transition: none;
    }
  }
</style>
