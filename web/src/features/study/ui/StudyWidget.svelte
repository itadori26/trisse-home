<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { formatEcts, formatGrade, type Progress } from '../domain/study';
  import type { StudyUseCases } from '../usecase/studyUseCases';

  /** Props of the study progress widget (#57). */
  interface Props {
    study: StudyUseCases;
  }
  let { study }: Props = $props();
  const poll = usePolling();

  // Local data that only changes in the settings; a look every few minutes keeps other tabs in sync.
  const POLL_MS = 5 * 60_000;
  let progress = $state<Progress | null>(null);
  let error = $state('');

  $effect(() =>
    poll(
      () =>
        study
          .progress()
          .then((p) => ((progress = p), (error = '')))
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      POLL_MS,
    ),
  );

  const ects = (v: number) => formatEcts(v, intlLocale());
</script>

{#if error && !progress}
  <p class="error">{error}</p>
{:else if progress}
  {#if progress.modules === 0}
    <p class="muted">{t('study.set_up')}</p>
  {:else}
    {@const p = progress}
    <div class="head">
      <strong>{t('study.of_ects', { passed: ects(p.passedEcts), total: ects(p.totalEcts) })}</strong>
      <span class="muted">{t('study.percent', { n: Math.round(p.percent) })}</span>
    </div>
    <div
      class="bar"
      role="progressbar"
      aria-valuemin="0"
      aria-valuemax={p.totalEcts}
      aria-valuenow={p.passedEcts}
      aria-valuetext={t('study.of_ects', { passed: ects(p.passedEcts), total: ects(p.totalEcts) })}
    >
      <span class="passed" style:width="{p.percent}%"></span>
      <span class="running" style:width="{Math.min(100 - p.percent, (p.inProgressEcts / p.totalEcts) * 100)}%"
      ></span>
    </div>
    <p class="muted small">{t('study.passed_modules', { n: p.passedModules, total: p.modules })}</p>
    {#if p.average !== null}
      <p class="average" title={t('study.average_hint', { ects: ects(p.gradedEcts) })}>
        {t('study.average', { grade: formatGrade(p.average, intlLocale()) })}
      </p>
    {/if}
    {#if p.inProgress.length}
      <h4>{t('study.in_progress')}</h4>
      <ul>
        {#each p.inProgress as m (m.id)}
          <li><span class="name">{m.name}</span><span class="muted">{ects(m.ects)} ECTS</span></li>
        {/each}
      </ul>
    {/if}
  {/if}
{/if}

<style>
  .head {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: var(--space-2);
  }
  .bar {
    display: flex;
    height: 0.8em;
    margin: var(--space-1) 0;
    border-radius: 999px;
    overflow: hidden;
    background: var(--color-surface-2);
  }
  .passed {
    background: var(--color-accent);
  }
  .running {
    background: color-mix(in srgb, var(--color-accent) 35%, transparent);
  }
  p {
    margin: 0;
  }
  .average {
    margin-top: var(--space-1);
    font-size: 1.1em;
    font-weight: 600;
  }
  h4 {
    margin: var(--space-2) 0 var(--space-1);
    font-size: 0.8em;
    color: var(--color-text-muted);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    justify-content: space-between;
    gap: var(--space-2);
    font-size: 0.9em;
  }
  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .small {
    font-size: 0.8em;
  }
</style>
