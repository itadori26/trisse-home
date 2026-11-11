<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { ShoppingSuggestion } from '../domain/shopping';
  import type { ShoppingUseCases } from '../usecase/shoppingUseCases';

  /** Props of the shopping list settings. */
  interface Props {
    shopping: ShoppingUseCases;
    /** Optional prefill, e.g. derived from the calendar settings. */
    suggest?: () => Promise<ShoppingSuggestion | null>;
  }
  let { shopping, suggest }: Props = $props();

  let listUrl = $state('');
  let username = $state('');
  let password = $state('');
  let hasPassword = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    shopping.settings().then(async (s) => {
      listUrl = s.listUrl;
      username = s.username;
      hasPassword = s.hasPassword;
      if (!s.configured && suggest) {
        const hint = await suggest().catch(() => null);
        if (hint) {
          listUrl ||= hint.listUrl;
          username ||= hint.username;
        }
      }
    });
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await shopping.saveSettings(listUrl, username, password);
      hasPassword = s.hasPassword;
      password = '';
      message = t('common.connected_saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<p class="muted">
  {t('shopping.the_shopping_list_is_stored_as')}
</p>
<form onsubmit={submit}>
  <label class="field"
    ><span>{t('shopping.task_list_url')}</span><input
      class="input"
      type="url"
      bind:value={listUrl}
      placeholder="https://cloud.example.com/remote.php/dav/calendars/USER/shopping/"
      required
    /></label
  >
  <label class="field"
    ><span>{t('shopping.username')}</span><input
      class="input"
      bind:value={username}
      autocomplete="off"
      required
    /></label
  >
  <label class="field">
    <span
      >{t('shopping.app_password', {
        value: hasPassword ? t('common.keep_stored') : '',
      })}</span
    >
    <input
      class="input"
      type="password"
      bind:value={password}
      autocomplete="new-password"
      required={!hasPassword}
    />
  </label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}
    >{busy ? t('common.testing') : t('common.test_save')}</button
  >
</form>
