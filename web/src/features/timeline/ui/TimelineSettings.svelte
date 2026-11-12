<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { TimelineUseCases } from '../usecase/timelineUseCases';

  /** Props of the timeline retention setting. */
  interface Props {
    timeline: TimelineUseCases;
  }
  let { timeline }: Props = $props();

  let days = $state(30);
  let message = $state('');
  let error = $state('');

  $effect(() => {
    timeline
      .retentionDays()
      .then((d) => (days = d))
      .catch((err) => (error = errorMessage(err)));
  });

  async function save(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    try {
      days = await timeline.setRetentionDays(days);
      message = t('common.saved');
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<form onsubmit={save}>
  <p class="muted">{t('timeline.settings_hint')}</p>
  <label class="field"
    ><span>{t('timeline.retention')}</span><input
      class="input"
      type="number"
      min="1"
      max="365"
      bind:value={days}
    /></label
  >
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit">{t('common.save')}</button>
</form>
