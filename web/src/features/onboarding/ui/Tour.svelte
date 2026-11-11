<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { availableSteps, type Box, popoverPosition, tourSteps } from '../domain/tour';

  /** Props of the guided tour overlay. */
  interface Props {
    /** Called when the tour is finished or skipped. */
    onfinish: () => void;
  }
  let { onfinish }: Props = $props();

  const PADDING = 6;
  const findTarget = (name: string) =>
    document.querySelector<HTMLElement>(`[data-tour="${CSS.escape(name)}"]`);

  const steps = availableSteps(tourSteps, (t) => findTarget(t) !== null);
  let index = $state(0);
  let box = $state<Box | null>(null);
  let pos = $state({ top: 0, left: 0 });
  let popover: HTMLElement | undefined = $state();

  const step = $derived(steps[index]);
  const last = $derived(index === steps.length - 1);
  const spotlightStyle = $derived(
    box ? `top:${box.top}px;left:${box.left}px;width:${box.width}px;height:${box.height}px` : '',
  );

  function measure() {
    const el = step?.target ? findTarget(step.target) : null;
    if (el) {
      el.scrollIntoView({ block: 'nearest' });
      const r = el.getBoundingClientRect();
      box = {
        top: r.top - PADDING,
        left: r.left - PADDING,
        width: r.width + 2 * PADDING,
        height: r.height + 2 * PADDING,
      };
    } else {
      box = null;
    }
    const size = popover
      ? { width: popover.offsetWidth, height: popover.offsetHeight }
      : { width: 360, height: 200 };
    pos = popoverPosition(box, size, { width: window.innerWidth, height: window.innerHeight });
  }

  $effect(() => {
    void index;
    void popover;
    measure();
    popover?.focus();
  });

  function next() {
    if (last) onfinish();
    else index++;
  }

  function back() {
    if (index > 0) index--;
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onfinish();
    else if (e.key === 'ArrowRight') next();
    else if (e.key === 'ArrowLeft') back();
  }
</script>

<svelte:window onresize={measure} {onkeydown} />

{#if step}
  <div class="tour" role="presentation">
    {#if box}
      <div class="spotlight" style={spotlightStyle}></div>
    {:else}
      <div class="backdrop"></div>
    {/if}
    <div
      class="popover card"
      bind:this={popover}
      role="dialog"
      aria-modal="true"
      aria-labelledby="tour-title"
      tabindex="-1"
      style:top="{pos.top}px"
      style:left="{pos.left}px"
    >
      <p class="progress muted">{index + 1} / {steps.length}</p>
      <h2 id="tour-title">{t(step.title)}</h2>
      <p>{t(step.body)}</p>
      <div class="actions">
        {#if !last}<button class="btn btn-icon skip" type="button" onclick={onfinish}
            >{t('onboarding.skip_tour')}</button
          >{/if}
        {#if index > 0}<button class="btn" type="button" onclick={back}>{t('onboarding.back')}</button>{/if}
        <button class="btn btn-primary" type="button" onclick={next}
          >{last ? t('tour.lets_go') : t('tour.next')}</button
        >
      </div>
    </div>
  </div>
{/if}

<style>
  .tour {
    position: fixed;
    inset: 0;
    z-index: 1000;
  }
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgb(0 0 0 / 55%);
  }
  /* The huge shadow darkens everything except the highlighted element. */
  .spotlight {
    position: fixed;
    border-radius: var(--radius);
    box-shadow: 0 0 0 9999px rgb(0 0 0 / 55%);
    outline: 2px solid var(--color-accent);
    transition: all 200ms ease;
    pointer-events: none;
  }
  .popover {
    position: fixed;
    width: min(380px, calc(100vw - 24px));
    padding: var(--space-4);
    outline: none;
    transition:
      top 200ms ease,
      left 200ms ease;
  }
  h2 {
    margin: 0 0 var(--space-2);
    font-size: 1.15rem;
  }
  .progress {
    margin: 0;
    font-size: 0.8rem;
  }
  .actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
    align-items: center;
  }
  .skip {
    margin-right: auto;
    color: var(--color-text-muted);
  }
</style>
