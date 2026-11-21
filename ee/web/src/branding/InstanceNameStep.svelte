<script lang="ts">
  import type { PluginContext } from '../../../../web/src/app/plugins/api';
  import { text } from '../lib/texts';
  import { BrandingApi, type BrandingSettings } from './branding';
  import { texts } from './texts';

  /**
   * The instance name as a step of the first-start flow (#671). The core
   * draws the buttons; Next saves the name with the other branding settings
   * unchanged.
   */
  interface Props {
    context: PluginContext;
    setSubmit: (save: () => Promise<void>) => void;
  }
  let { context, setSubmit }: Props = $props();

  const api = $derived(new BrandingApi(context.http));
  const tr = (key: string) => text(texts, key, context.language);

  let settings = $state<BrandingSettings | null>(null);
  let name = $state('');
  let error = $state('');

  $effect(() => {
    api
      .settings()
      .then((s) => {
        settings = s;
        name = s.name;
      })
      .catch((err) => (error = context.errorMessage(err)));
  });

  $effect(() => {
    setSubmit(async () => {
      if (!settings || name.trim() === settings.name) return;
      settings = await api.save({ ...settings, name });
    });
  });
</script>

<label class="field">
  <span>{tr('step_name_label')}</span>
  <input class="input" bind:value={name} maxlength="60" placeholder="Waylight" />
</label>
<p class="muted small">{tr('step_name_hint')}</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .small {
    font-size: 0.85rem;
  }
</style>
