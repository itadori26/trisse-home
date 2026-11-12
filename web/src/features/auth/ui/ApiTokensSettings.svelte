<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import CopySnippet from '../../../shared/ui/CopySnippet.svelte';
  import { type Access, type ApiToken, scopeModules, toScopes } from '../domain/apiToken';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** Props of the API token section (owner only). */
  interface Props {
    auth: AuthUseCases;
    modules: Set<string>;
  }
  let { auth, modules }: Props = $props();

  let tokens = $state<ApiToken[]>([]);
  let error = $state('');
  let name = $state('');
  let days = $state(0);
  let access = $state<Record<string, Access>>({});
  let created = $state('');
  const choices = $derived(scopeModules(modules));

  const load = () =>
    auth
      .apiTokens()
      .then((l) => (tokens = l))
      .catch((err) => (error = errorMessage(err)));

  $effect(() => {
    load();
  });

  async function create(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    try {
      created = await auth.createApiToken(name, toScopes(access), days);
      name = '';
      access = {};
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function revoke(tok: ApiToken) {
    if (!confirm(t('auth.token_revoke_confirm', { name: tok.name }))) return;
    try {
      await auth.revokeApiToken(tok.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const date = (d: Date | null) => (d ? d.toLocaleString(intlLocale()) : t('auth.token_never'));
</script>

<p class="muted">{t('auth.tokens_hint')}</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if created}
  <div class="created" role="status">
    <p><strong>{t('auth.token_created')}</strong></p>
    <CopySnippet code={created} label={t('auth.new_token')} />
    <button class="btn" type="button" onclick={() => (created = '')}>{t('auth.token_saved_it')}</button>
  </div>
{/if}

<ul class="tokens">
  {#each tokens as tok (tok.id)}
    <li>
      <div>
        <strong>{tok.name}</strong> <code class="small">{tok.scopes.join(' ')}</code>
        <div class="muted small">
          {t('auth.token_last_used', { when: date(tok.lastUsedAt) })}{#if tok.lastUsedIp}
            · {tok.lastUsedIp}{/if} ·
          {t('auth.token_expires', { when: date(tok.expiresAt) })}
        </div>
      </div>
      <button class="btn btn-danger" type="button" onclick={() => revoke(tok)}
        >{t('auth.token_revoke')}</button
      >
    </li>
  {:else}
    <li class="muted">{t('auth.no_tokens')}</li>
  {/each}
</ul>

<form onsubmit={create}>
  <h3>{t('auth.new_token')}</h3>
  <label class="field"
    ><span>{t('auth.token_name')}</span><input
      class="input"
      bind:value={name}
      maxlength="60"
      required
    /></label
  >
  <fieldset class="scopes">
    <legend>{t('auth.token_scopes')}</legend>
    {#each choices as m (m)}
      <label
        ><span>{m}</span>
        <select
          class="input"
          value={access[m] ?? 'none'}
          onchange={(e) => (access[m] = e.currentTarget.value as Access)}
        >
          <option value="none">{t('auth.access_none')}</option>
          <option value="read">{t('auth.access_read')}</option>
          <option value="write">{t('auth.access_write')}</option>
        </select></label
      >
    {/each}
  </fieldset>
  <label class="field"
    ><span>{t('auth.token_expiry_days')}</span><input
      class="input"
      type="number"
      min="0"
      max="3650"
      bind:value={days}
    /></label
  >
  <button class="btn btn-primary" type="submit">{t('auth.create_token')}</button>
</form>

<style>
  .tokens {
    list-style: none;
    padding: 0;
  }
  .tokens li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  .scopes {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(11rem, 1fr));
    gap: var(--space-1);
    border: 0;
    padding: 0;
  }
  .scopes label {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-1);
  }
  .created {
    padding: var(--space-2);
    border-radius: var(--radius-small);
    background: color-mix(in srgb, var(--color-accent) 15%, transparent);
  }
  .small {
    font-size: 0.85em;
  }
</style>
