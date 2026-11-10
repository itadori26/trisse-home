<script lang="ts" generics="T extends LayoutItem">
  import type { Snippet } from 'svelte';
  import { GRID_COLUMNS, type LayoutItem, moveItem, resizeItem } from '../domain/layout';

  /** Props of the drag & drop grid. */
  interface Props {
    /** Items to render; their x/y/w/h are grid cells. */
    items: T[];
    /** Enables dragging and resizing. */
    editing: boolean;
    /** Height of one grid row in pixels. */
    rowHeight?: number;
    /** Gap between cells in pixels. */
    gap?: number;
    /** Called once after a drag or resize with the new layout. */
    onlayout: (items: LayoutItem[]) => void;
    /** Renders the content of one item; `handle` must be spread onto the drag handle. */
    item: Snippet<[T, { onpointerdown: (e: PointerEvent) => void }]>;
  }
  let { items, editing, rowHeight = 56, gap = 12, onlayout, item }: Props = $props();

  let container: HTMLDivElement;
  let preview = $state<LayoutItem[] | null>(null);

  type Gesture = { mode: 'move' | 'resize'; id: string; startX: number; startY: number; origin: LayoutItem };
  let gesture = $state<Gesture | null>(null);

  const rects = $derived(new Map((preview ?? items).map((i) => [i.id, i])));

  function cellSize() {
    const width = container.clientWidth;
    return { col: (width - gap * (GRID_COLUMNS - 1)) / GRID_COLUMNS + gap, row: rowHeight + gap };
  }

  function start(mode: Gesture['mode'], target: T, e: PointerEvent) {
    if (!editing || e.button !== 0) return;
    e.preventDefault();
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    gesture = { mode, id: target.id, startX: e.clientX, startY: e.clientY, origin: { ...target } };
    preview = items.map(({ id, x, y, w, h }) => ({ id, x, y, w, h }));
  }

  function move(e: PointerEvent) {
    if (!gesture) return;
    const { col, row } = cellSize();
    const dx = Math.round((e.clientX - gesture.startX) / col);
    const dy = Math.round((e.clientY - gesture.startY) / row);
    const base = items.map(({ id, x, y, w, h }) => ({ id, x, y, w, h }));
    const o = gesture.origin;
    preview =
      gesture.mode === 'move'
        ? moveItem(base, o.id, o.x + dx, o.y + dy)
        : resizeItem(base, o.id, o.w + dx, o.h + dy);
  }

  function end() {
    if (!gesture) return;
    gesture = null;
    if (preview) onlayout(preview);
    preview = null;
  }
</script>

<div
  class="grid"
  class:editing
  bind:this={container}
  style:--row-height="{rowHeight}px"
  style:--gap="{gap}px"
  onpointermove={move}
  onpointerup={end}
  onpointercancel={end}
  role="presentation"
>
  {#each items as it (it.id)}
    {@const r = rects.get(it.id) ?? it}
    <div
      class="cell"
      class:active={preview && gesture?.id === it.id}
      style:grid-area="{r.y + 1} / {r.x + 1} / span {r.h} / span {r.w}"
      style:--span={r.h}
    >
      {@render item(it, { onpointerdown: (e) => start('move', it, e) })}
      {#if editing}
        <button
          class="resize"
          type="button"
          aria-label="Resize widget"
          onpointerdown={(e) => start('resize', it, e)}
        ></button>
      {/if}
    </div>
  {/each}
</div>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(12, minmax(0, 1fr));
    grid-auto-rows: var(--row-height);
    gap: var(--gap);
    min-height: 200px;
  }
  .grid.editing {
    background-image: linear-gradient(to right, transparent calc(100% - 1px), var(--color-border) 0);
    background-size: calc((100% + var(--gap)) / 12) 100%;
    touch-action: none;
    user-select: none;
  }
  .cell {
    position: relative;
    min-width: 0;
    min-height: 0;
  }
  .cell.active {
    z-index: 2;
    opacity: 0.85;
  }
  .resize {
    position: absolute;
    right: 2px;
    bottom: 2px;
    width: 18px;
    height: 18px;
    border: none;
    background: linear-gradient(135deg, transparent 50%, var(--color-accent) 50%);
    border-radius: 0 0 var(--radius) 0;
    cursor: nwse-resize;
  }
  /* On phones widgets are stacked and sized by their content: the desktop
     height is the minimum, long lists are capped and scroll. */
  @media (max-width: 720px) {
    .grid {
      grid-template-columns: 1fr;
      grid-auto-rows: auto;
    }
    .cell {
      grid-area: auto !important;
      min-height: calc(var(--span, 4) * (var(--row-height) + var(--gap)) - var(--gap));
      max-height: 70vh;
      display: flex;
    }
    .cell > :global(*) {
      flex: 1;
    }
  }
</style>
