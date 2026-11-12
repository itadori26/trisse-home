<script lang="ts">
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { formatBytes } from '../../../shared/ui/format';
  import { type Disk, type Finding, isKnownFinding, overallStatus, powerOnAge } from '../domain/disk';
  import type { DiskHealthUseCases } from '../usecase/diskHealthUseCases';

  /** Props of the disk health widget (#59). */
  interface Props {
    diskhealth: DiskHealthUseCases;
  }
  let { diskhealth }: Props = $props();

  let disks = $state<Disk[] | null>(null);
  let error = $state('');
  let missing = $state(false);

  $effect(() => {
    const load = () =>
      diskhealth
        .disks()
        .then((d) => ((disks = d), (error = ''), (missing = false)))
        .catch((err) => {
          missing = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
        });
    load();
    const timer = setInterval(load, 5 * 60_000);
    return () => clearInterval(timer);
  });

  const describe = (f: Finding) =>
    isKnownFinding(f.code) ? t(`diskhealth.finding.${f.code}`, { value: f.value, name: f.name }) : f.message;
  const age = (h: number) => {
    const a = powerOnAge(h);
    return t('diskhealth.age', { years: a.years, days: a.days });
  };
</script>

{#if error && !disks}
  <p class={missing ? 'muted' : 'error'}>{error}</p>
{:else if disks}
  <p class="summary status-{overallStatus(disks)}">{t(`diskhealth.overall.${overallStatus(disks)}`)}</p>
  <ul class="disks">
    {#each disks as d (d.device)}
      <li class="disk status-{d.status}">
        <div class="head">
          <span class="dot" aria-hidden="true"></span>
          <strong>{d.model || d.device}</strong>
          <span class="muted small"
            >{d.device}{d.capacityBytes ? ` · ${formatBytes(d.capacityBytes)}` : ''}</span
          >
          <span class="badge">{t(`diskhealth.status.${d.status}`)}</span>
        </div>
        <dl class="facts small">
          {#if d.temperature !== null}<dt>{t('diskhealth.temperature')}</dt>
            <dd>{d.temperature} °C</dd>{/if}
          {#if d.reallocated !== null}<dt>{t('diskhealth.reallocated')}</dt>
            <dd>{d.reallocated}</dd>{/if}
          {#if d.pending !== null}<dt>{t('diskhealth.pending')}</dt>
            <dd>{d.pending}</dd>{/if}
          {#if d.percentageUsed !== null}<dt>{t('diskhealth.wear')}</dt>
            <dd>{d.percentageUsed} %</dd>{/if}
          {#if d.powerOnHours !== null}<dt>{t('diskhealth.power_on')}</dt>
            <dd title={t('diskhealth.hours', { hours: d.powerOnHours })}>{age(d.powerOnHours)}</dd>{/if}
        </dl>
        {#if d.findings.length}
          <ul class="findings small">
            {#each d.findings as f, i (i)}<li class="status-{f.status}">{describe(f)}</li>{/each}
          </ul>
        {/if}
      </li>
    {:else}
      <li class="muted">{t('diskhealth.no_disks')}</li>
    {/each}
  </ul>
{:else}
  <p class="muted">{t('diskhealth.loading')}</p>
{/if}

<style>
  .disks,
  .findings {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  .disk {
    padding: var(--space-1) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  .head {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    flex-wrap: wrap;
  }
  .badge {
    margin-left: auto;
    font-size: 0.8em;
  }
  .dot {
    width: 0.6em;
    height: 0.6em;
    border-radius: 50%;
    background: var(--color-muted, #888);
  }
  .status-ok .dot {
    background: var(--color-success);
  }
  .status-warning .dot {
    background: var(--color-warning);
  }
  .status-critical .dot {
    background: var(--color-danger);
  }
  .facts {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0 var(--space-2);
    margin: var(--space-1) 0 0;
  }
  .facts dt {
    color: var(--color-muted, #888);
  }
  .facts dd {
    margin: 0;
  }
  .findings li.status-warning,
  .summary.status-warning {
    color: var(--color-warning);
  }
  .findings li.status-critical,
  .summary.status-critical {
    color: var(--color-danger);
  }
</style>
