<script lang="ts">
  import { t } from '../../../../shared/i18n';
  import Meter from '../../../../shared/ui/Meter.svelte';
  import { type ActionItem, type Answer, dataAge } from '../../domain/smart/answer';

  /** Props of one answer of the smart search (#156); generic per answer kind. */
  interface Props {
    answer: Answer;
    /** Opens the widget with the details, when the dashboard has one. */
    onopen?: (widgetKind: string) => void;
    /** Runs a confirmed action. */
    onexecute?: (item: ActionItem) => void;
    /** Puts a suggestion into the search field (#158). */
    onpick?: (text: string) => void;
  }
  let { answer, onopen, onexecute, onpick }: Props = $props();

  const age = $derived(dataAge(answer.at, new Date()));
</script>

<article class="answer" data-kind={answer.kind} aria-label={answer.title}>
  <header>
    <h3>{answer.title}</h3>
    {#if age}<span class="age">{t(age.key, { n: age.n })}</span>{/if}
  </header>

  {#if answer.kind === 'metric'}
    <p class="value" data-severity={answer.severity}>{answer.value}</p>
    {#if answer.percent !== undefined}<Meter percent={answer.percent} label={answer.title} />{/if}
    {#if answer.detail}<p class="detail">{answer.detail}</p>{/if}
  {:else if answer.kind === 'status'}
    <p class="status"><span class="dot" data-state={answer.state} aria-hidden="true"></span>{answer.text}</p>
    {#if answer.detail}<p class="detail">{answer.detail}</p>{/if}
  {:else if answer.kind === 'list'}
    {#if answer.items.length === 0}
      <p class="detail">{answer.empty}</p>
    {:else}
      <ul>
        {#each answer.items as item, i (i)}
          <li>
            {#if item.state}<span class="dot" data-state={item.state} aria-hidden="true"></span>{/if}
            <span class="label">{item.label}</span>
            <span class="detail">{item.value}</span>
          </li>
        {/each}
      </ul>
    {/if}
  {:else if answer.kind === 'text'}
    <p>{answer.text}</p>
    {#if answer.suggestions?.length}
      <p class="chips">
        {#each answer.suggestions as s (s)}
          <button type="button" class="btn chip" onclick={() => onpick?.(s)}>{s}</button>
        {/each}
      </p>
    {/if}
  {:else if answer.kind === 'actions'}
    <p class="detail">{answer.text}</p>
    <p class="chips">
      {#each answer.actions as a (a.action)}
        <button type="button" class="btn btn-primary" onclick={() => confirm(a.confirm) && onexecute?.(a)}
          >{a.label}</button
        >
      {/each}
    </p>
  {/if}

  {#if answer.widgetKind && onopen}
    <button type="button" class="btn btn-icon open" onclick={() => onopen?.(answer.widgetKind ?? '')}
      >{t('smart.open_widget')}</button
    >
  {/if}
</article>

<style>
  .answer {
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
    background: var(--color-surface);
    display: grid;
    gap: var(--space-1);
  }
  header {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: var(--space-2);
  }
  h3 {
    margin: 0;
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--color-text-muted);
  }
  p {
    margin: 0;
  }
  .age,
  .detail {
    font-size: 0.8rem;
    color: var(--color-text-muted);
  }
  .value {
    font-size: 1.5rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  .value[data-severity='warning'] {
    color: var(--color-warning);
  }
  .value[data-severity='critical'] {
    color: var(--color-danger);
  }
  .status {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-weight: 500;
  }
  .dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    flex: none;
    background: var(--color-text-muted);
  }
  .dot[data-state='ok'] {
    background: var(--color-success, #2e9d55);
  }
  .dot[data-state='warning'] {
    background: var(--color-warning);
  }
  .dot[data-state='error'] {
    background: var(--color-danger);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 2px;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }
  .chip {
    padding: 2px var(--space-2);
  }
  .open {
    justify-self: start;
    font-size: 0.8rem;
    padding: 0;
    color: var(--color-accent);
  }
</style>
