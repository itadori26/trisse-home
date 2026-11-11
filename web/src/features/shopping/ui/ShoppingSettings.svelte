<script lang="ts">
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
      message = 'Connected and saved.';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<p class="muted">
  The shopping list is stored as tasks in a CalDAV task list, so everyone in your household can use it in
  their task app (e.g. Nextcloud Tasks, Apple Reminders, DAVx⁵ + Tasks.org). Create a list called “shopping”
  in Nextcloud Tasks and use its URL.
</p>
<form onsubmit={submit}>
  <label class="field"
    ><span>Task list URL</span><input
      class="input"
      type="url"
      bind:value={listUrl}
      placeholder="https://cloud.example.com/remote.php/dav/calendars/USER/shopping/"
      required
    /></label
  >
  <label class="field"
    ><span>Username</span><input class="input" bind:value={username} autocomplete="off" required /></label
  >
  <label class="field">
    <span>App password {hasPassword ? '(leave empty to keep the stored one)' : ''}</span>
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
  <button class="btn btn-primary" type="submit" disabled={busy}>{busy ? 'Testing…' : 'Test & save'}</button>
</form>
