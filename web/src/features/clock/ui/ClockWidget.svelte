<script lang="ts">
  import { formatDate, formatTime, parseClockConfig } from '../domain/clock';

  /** Props of the clock widget. */
  interface Props {
    config: Record<string, unknown>;
  }
  let { config }: Props = $props();

  const cfg = $derived(parseClockConfig(config));
  let now = $state(new Date());

  $effect(() => {
    const interval = cfg.showSeconds ? 1000 : 10_000;
    const timer = setInterval(() => (now = new Date()), interval);
    return () => clearInterval(timer);
  });
</script>

<div class="clock">
  <time class="time" datetime={now.toISOString()}>{formatTime(now, cfg)}</time>
  {#if cfg.showDate}<div class="date">{formatDate(now, cfg)}</div>{/if}
  {#if cfg.label}<div class="label muted">{cfg.label}</div>{/if}
</div>

<style>
  .clock {
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    text-align: center;
  }
  .time {
    font-size: clamp(1.6rem, min(19cqi, 42cqh), 10rem);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    line-height: 1.1;
  }
  .date {
    color: var(--color-text-muted);
    font-size: clamp(0.8rem, min(5.5cqi, 9cqh), 2rem);
  }
  .label {
    font-size: clamp(0.75rem, min(4cqi, 7cqh), 1.4rem);
  }
</style>
