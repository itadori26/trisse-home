<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import CopySnippet from '../../../shared/ui/CopySnippet.svelte';
  import type { Peer } from '../domain/peer';
  import { exportCronSnippet } from '../domain/snippet';
  import type { WireGuardUseCases } from '../usecase/wireguardUseCases';

  /** Props of the WireGuard peer names section. */
  interface Props {
    wireguard: WireGuardUseCases;
  }
  let { wireguard }: Props = $props();

  let peers = $state<Peer[]>([]);
  let names = $state<Record<string, string>>({});
  let error = $state('');
  let saved = $state('');
  let stackDir = $state('/opt/stacks/trisse-home');
  let container = $state('wireguard');
  const snippet = $derived(exportCronSnippet({ stackDir, container }));

  $effect(() => {
    wireguard
      .peers()
      .then((l) => {
        peers = l.peers;
        names = Object.fromEntries(l.peers.map((p) => [p.publicKey, p.name.endsWith('…') ? '' : p.name]));
      })
      .catch((err) => (error = errorMessage(err)));
  });

  async function save(p: Peer) {
    error = saved = '';
    try {
      await wireguard.rename(p.publicKey, names[p.publicKey] ?? '');
      saved = p.publicKey;
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<p class="muted">
  {t('wireguard.the_wireguard_widget_reads_a_file')} <code>scripts/wireguard-export.sh</code>
  {t('wireguard.on_the_host_see_readme_names')}
</p>
<details>
  <summary>{t('wireguard.setup_snippet')}</summary>
  <label class="field"
    ><span>{t('wireguard.stack_dir')}</span><input
      class="input"
      bind:value={stackDir}
      autocomplete="off"
    /></label
  >
  <label class="field"
    ><span>{t('wireguard.container')}</span><input
      class="input"
      bind:value={container}
      placeholder={t('wireguard.container_host')}
      autocomplete="off"
    /></label
  >
  {#if snippet}
    <CopySnippet code={snippet} label={t('wireguard.setup_snippet')} />
  {:else}
    <p class="error">{t('wireguard.snippet_invalid')}</p>
  {/if}
  <p class="muted">
    <a
      href="https://github.com/itadori26/trisse-home/blob/main/docs/security-model.md#wireguard-export"
      target="_blank"
      rel="noopener noreferrer">{t('common.why_extra_step')}</a
    >
  </p>
</details>
{#if error}<p class={peers.length ? 'error' : 'muted'} role="alert">{error}</p>{/if}
<ul class="list">
  {#each peers as p (p.publicKey)}
    <li>
      <code title={p.publicKey}>{p.publicKey.slice(0, 10)}…</code>
      <input
        class="input"
        bind:value={names[p.publicKey]}
        placeholder={t('wireguard.name')}
        maxlength="40"
        aria-label={t('wireguard.name_of_peer', { value: p.publicKey.slice(0, 10) })}
        onchange={() => save(p)}
      />
      {#if saved === p.publicKey}<span class="muted">✓</span>{/if}
    </li>
  {/each}
</ul>

<style>
  .list {
    list-style: none;
    padding: 0;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-1);
  }
  .input {
    flex: 1;
  }
</style>
