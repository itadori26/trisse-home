<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import type { CalendarUseCases } from '../usecase/calendarUseCases';

  /** Props of the CalDAV connection settings. */
  interface Props {
    calendar: CalendarUseCases;
  }
  let { calendar }: Props = $props();

  let urls = $state('');
  let username = $state('');
  let password = $state('');
  let hasPassword = $state(false);
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    calendar
      .settings()
      .then(
        (s) => ((urls = s.calendarUrls.join('\n')), (username = s.username), (hasPassword = s.hasPassword)),
      );
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await calendar.saveSettings(urls, username, password);
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

<form onsubmit={submit}>
  <label class="field">
    <span>Calendar URLs (one per line)</span>
    <textarea
      class="input"
      rows="3"
      bind:value={urls}
      required
      placeholder="https://cloud.example.com/remote.php/dav/calendars/USER/personal/"></textarea>
  </label>
  <p class="muted">
    Nextcloud: open the calendar app, click “…” next to a calendar → “Copy private link”, and use that URL.
    Create an app password under Settings → Security.
  </p>
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
