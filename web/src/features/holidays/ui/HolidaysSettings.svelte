<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { HolidaysUseCases } from '../usecase/holidaysUseCases';

  /** Settings → Holidays: the optional ICS link of the school holidays (#55). */
  interface Props {
    holidays: HolidaysUseCases;
  }
  let { holidays }: Props = $props();

  let schoolUrl = $state('');
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    holidays.settings().then((s) => (schoolUrl = s.schoolUrl));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await holidays.saveSettings(schoolUrl);
      schoolUrl = s.schoolUrl;
      message = s.schoolUrl ? t('holidays.saved') : t('holidays.school_removed');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<p class="muted">{t('holidays.school_hint')}</p>
<form onsubmit={submit}>
  <label class="field"
    ><span>{t('holidays.school_url')}</span><input
      class="input"
      bind:value={schoolUrl}
      placeholder="https://…/schulferien.ics"
    /></label
  >
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted" role="status">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.loading') : t('common.test_save')}</button
  >
</form>
