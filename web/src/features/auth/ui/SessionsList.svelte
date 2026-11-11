<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { describeUserAgent, type SessionInfo } from '../domain/session';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** Props of the session overview. */
  interface Props {
    auth: AuthUseCases;
  }
  let { auth }: Props = $props();

  let sessions = $state<SessionInfo[]>([]);
  let error = $state('');
  let message = $state('');

  async function run(action: () => Promise<unknown>) {
    error = message = '';
    try {
      await action();
      sessions = await auth.sessions();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    run(async () => {});
  });
</script>

<ul class="list">
  {#each sessions as s (s.id)}
    <li>
      <span class="who">
        <strong>{describeUserAgent(s.userAgent)}</strong>
        {#if s.current}<span class="badge">this device</span>{/if}
        <br /><span class="muted small"
          >{s.clientIp || 'unknown IP'} · last active {s.lastSeenAt.toLocaleString()}</span
        >
      </span>
      {#if !s.current}
        <button
          class="btn btn-icon"
          type="button"
          aria-label="Sign out this session"
          onclick={() => run(() => auth.revokeSession(s.id))}>Sign out</button
        >
      {/if}
    </li>
  {/each}
</ul>
{#if sessions.length > 1}
  <button
    class="btn"
    type="button"
    onclick={() =>
      run(async () => {
        const n = await auth.revokeOtherSessions();
        message = `${n} other session(s) signed out.`;
      })}>Sign out all other sessions</button
  >
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted">{message}</p>{/if}

<style>
  .list {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-2);
  }
  .list li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .who {
    flex: 1;
  }
  .small {
    font-size: 0.8rem;
  }
  .badge {
    margin-left: var(--space-2);
    padding: 0 6px;
    border-radius: 999px;
    background: var(--color-accent);
    color: var(--color-accent-contrast);
    font-size: 0.75rem;
  }
</style>
