<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { GameDealsSettings } from '../domain/gamedeals';
  import type { GamedealsUseCases } from '../usecase/gamedealsUseCases';

  /** Props of the game deals settings. */
  interface Props {
    gamedeals: GamedealsUseCases;
  }
  let { gamedeals }: Props = $props();

  let form = $state<GameDealsSettings>({ steamId: '', country: 'DE', minDiscount: 20, epic: false });
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  $effect(() => {
    gamedeals
      .settings()
      .then((s) => (form = s))
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      form = await gamedeals.saveSettings(form);
      message = t('common.saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<form onsubmit={submit}>
  <label class="field"
    ><span>{t('gamedeals.steam_id')}</span><input
      class="input"
      bind:value={form.steamId}
      placeholder="https://steamcommunity.com/profiles/7656119…"
      autocomplete="off"
    /></label
  >
  <p class="muted">{t('gamedeals.steam_hint')}</p>
  <div class="row">
    <label class="field"
      ><span>{t('gamedeals.country')}</span><input
        class="input"
        bind:value={form.country}
        maxlength="2"
      /></label
    >
    <label class="field"
      ><span>{t('gamedeals.min_discount')}</span><input
        class="input"
        type="number"
        min="0"
        max="100"
        bind:value={form.minDiscount}
      /></label
    >
  </div>
  <label><input type="checkbox" bind:checked={form.epic} /> {t('gamedeals.epic')}</label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted">{message}</p>{/if}
  <p><button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button></p>
</form>

<style>
  .row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-2);
  }
</style>
