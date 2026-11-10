<script lang="ts">
  /** Props of a horizontal usage bar. */
  interface Props {
    label: string;
    /** 0–100 */
    percent: number;
    /** Text shown on the right, e.g. "3.1 / 7.6 GiB". */
    detail?: string;
  }
  let { label, percent, detail = '' }: Props = $props();
  const level = $derived(percent >= 90 ? 'danger' : percent >= 75 ? 'warning' : 'ok');
</script>

<div class="meter">
  <div class="row">
    <span>{label}</span>
    <span class="muted">{detail || `${percent.toFixed(0)} %`}</span>
  </div>
  <div
    class="bar"
    role="meter"
    aria-label={label}
    aria-valuenow={percent}
    aria-valuemin="0"
    aria-valuemax="100"
  >
    <div class="fill {level}" style:width="{Math.min(100, Math.max(0, percent))}%"></div>
  </div>
</div>

<style>
  .meter {
    margin-bottom: var(--space-2);
  }
  .row {
    display: flex;
    justify-content: space-between;
    font-size: 0.85rem;
  }
  .bar {
    height: 6px;
    border-radius: 3px;
    background: var(--color-surface-2);
    overflow: hidden;
  }
  .fill {
    height: 100%;
    background: var(--color-accent);
    transition: width 400ms ease;
  }
  .fill.warning {
    background: var(--color-warning);
  }
  .fill.danger {
    background: var(--color-danger);
  }
</style>
