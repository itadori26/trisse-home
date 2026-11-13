<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import CopySnippet from '../../../shared/ui/CopySnippet.svelte';
  import { backgroundUrl } from '../domain/space';
  import type { SpaceUseCases } from '../usecase/spaceUseCases';

  /** Settings → Space (#64). */
  interface Props {
    space: SpaceUseCases;
  }
  let { space }: Props = $props();

  let launches = $state(false);
  let picture = $state(false);
  let hasKey = $state(false);
  let key = $state('');
  let error = $state('');
  let message = $state('');

  $effect(() => {
    space.settings().then((s) => ((launches = s.launches), (picture = s.picture), (hasKey = s.hasKey)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    try {
      const s = await space.saveSettings(launches, picture, key);
      hasKey = s.hasKey;
      key = '';
      message = t('space.saved');
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<p class="muted">{t('space.intro')}</p>
<form onsubmit={submit}>
  <label class="check"><input type="checkbox" bind:checked={launches} /> {t('space.launches_on')}</label>
  <label class="check"><input type="checkbox" bind:checked={picture} /> {t('space.picture_on')}</label>
  {#if picture}
    <label class="field"
      ><span>{t('space.nasa_key')}</span><input
        class="input"
        type="password"
        autocomplete="off"
        bind:value={key}
        placeholder={hasKey ? t('space.nasa_key_stored') : 'DEMO_KEY'}
      /></label
    >
    <p class="muted small">{t('space.nasa_key_hint')}</p>
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted" role="status">{message}</p>{/if}
  <button class="btn btn-primary" type="submit">{t('common.save')}</button>
</form>
{#if picture}
  <h3>{t('space.background')}</h3>
  <p class="muted small">{t('space.background_hint')}</p>
  <CopySnippet code={backgroundUrl(location.origin)} label={t('space.background_copy')} />
{/if}

<style>
  form {
    display: grid;
    gap: var(--space-2);
    justify-items: start;
  }
  form .field {
    justify-self: stretch;
  }
  .check {
    display: flex;
    gap: var(--space-1);
    align-items: center;
  }
  h3 {
    margin: var(--space-3) 0 var(--space-1);
    font-size: 1em;
  }
  .small {
    font-size: 0.85em;
    margin: 0;
  }
</style>
