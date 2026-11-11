<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { AuthUseCases } from '../usecase/authUseCases';
  import AuthCard from './AuthCard.svelte';

  /** Props of the first-run setup screen. */
  interface Props {
    auth: AuthUseCases;
    /** Called after the account was created and the user is logged in. */
    ondone: () => void;
  }
  let { auth, ondone }: Props = $props();

  let token = $state('');
  let password = $state('');
  let confirmation = $state('');
  let error = $state('');
  let busy = $state(false);

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    busy = true;
    error = '';
    try {
      await auth.setup(token, password, confirmation);
      ondone();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<AuthCard
  title={t('auth.welcome_to_trisse_home')}
  subtitle="Set a password to finish the setup. The setup token is printed in the server log."
>
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('auth.setup_token')}</span>
      <input class="input" name="setupToken" autocomplete="off" bind:value={token} required />
    </label>
    <label class="field">
      <span>{t('auth.password_at_least_12_characters')}</span>
      <input
        class="input"
        type="password"
        name="password"
        autocomplete="new-password"
        bind:value={password}
        required
      />
    </label>
    <label class="field">
      <span>{t('auth.repeat_password')}</span>
      <input
        class="input"
        type="password"
        name="confirmation"
        autocomplete="new-password"
        bind:value={confirmation}
        required
      />
    </label>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('auth.create_account')}</button>
  </form>
</AuthCard>
