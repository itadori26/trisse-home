<script lang="ts">
  import { t } from '../../../shared/i18n';
  import Modal from '../../../shared/ui/Modal.svelte';

  /** Props of a small dialog asking for one line of text. */
  interface Props {
    title: string;
    label: string;
    initial?: string;
    /** Allows an empty value (e.g. to reset a setting). */
    optional?: boolean;
    /** Input type and length limit. */
    type?: 'text' | 'url';
    maxlength?: number;
    placeholder?: string;
    /** Called with the entered text; may throw to show an error. */
    onsubmit: (value: string) => Promise<void>;
    onclose: () => void;
  }
  let {
    title,
    label,
    initial = '',
    optional = false,
    type = 'text',
    maxlength = 40,
    placeholder = '',
    onsubmit,
    onclose,
  }: Props = $props();

  // The dialog is recreated for every prompt, so capturing the initial value is intended.
  // svelte-ignore state_referenced_locally
  let value = $state(initial);
  let error = $state('');

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    try {
      await onsubmit(value);
      onclose();
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    }
  }
</script>

<Modal {title} {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>{label}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" {type} bind:value autofocus required={!optional} {maxlength} {placeholder} />
    </label>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit">{t('dashboard.save')}</button>
  </form>
</Modal>
