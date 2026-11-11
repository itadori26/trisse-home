<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { describeResult } from '../domain/transfer';
  import type { TransferUseCases } from '../usecase/transferUseCases';

  /** Props of the export/import section. */
  interface Props {
    transfer: TransferUseCases;
  }
  let { transfer }: Props = $props();

  let message = $state('');
  let error = $state('');
  let busy = $state(false);

  async function onfile(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    if (
      !confirm(
        'Importing replaces your dashboard, bookmarks, notes, to-dos, feeds, services and settings with the content of the file. Continue?',
      )
    )
      return;
    busy = true;
    error = message = '';
    try {
      const res = await transfer.importFile(await file.text());
      message = describeResult(res) + ' Reload the page to see everything.';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<h4>Export</h4>
<p class="muted">
  Downloads your layout, bookmarks, notes, to-dos, feeds, services and settings as JSON. Passwords, tokens and
  API keys are never included.
</p>
<p><a class="btn" href={transfer.exportUrl()} download>Download export</a></p>

<h4>Import</h4>
<p class="muted">
  Restores an export, e.g. on a new server. Credentials have to be entered again afterwards.
</p>
<label class="btn file">
  {busy ? 'Importing…' : 'Choose export file…'}
  <input type="file" accept="application/json,.json" onchange={onfile} disabled={busy} />
</label>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  h4 {
    margin: var(--space-3) 0 var(--space-1);
  }
  .file input {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
  }
  .file:focus-within {
    outline: 2px solid var(--color-accent);
  }
</style>
