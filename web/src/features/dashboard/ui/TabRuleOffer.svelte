<script lang="ts">
  import { t, type MessageKey } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { Tab, TabRule } from '../domain/dashboard';
  import { presetRule, rulePresets, type RulePreset } from '../domain/schedule';

  /**
   * Props of the one-time question after a tab was created (#673). It sits
   * inline instead of in a dialog so creating a tab is never interrupted.
   */
  interface Props {
    tab: Tab;
    /** The client's private network; without it "in this network" is not offered. */
    clientNetwork?: string;
    /** Saves the tab's rules; may throw to show an error. */
    onsubmit: (rules: TabRule[]) => Promise<void>;
    /** Opens the full rules dialog of the tab. */
    onedit: () => void;
    /** Ends the offer for good, answered or not. */
    onclose: () => void;
  }
  let { tab, clientNetwork = '', onsubmit, onedit, onclose }: Props = $props();

  const labels: Record<RulePreset, MessageKey> = {
    mornings: 'dashboard.offer_preset_mornings',
    workdays: 'dashboard.offer_preset_workdays',
    network: 'dashboard.offer_preset_network',
  };

  let saved = $state(false);
  let busy = $state(false);
  let error = $state('');

  async function choose(preset: RulePreset) {
    error = '';
    busy = true;
    try {
      await onsubmit([...tab.rules, presetRule(preset, clientNetwork)]);
      saved = true;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<div class="offer" role="group" aria-labelledby="tab-rule-offer" data-testid="tab-rule-offer">
  {#if saved}
    <p id="tab-rule-offer" role="status">{t('dashboard.offer_saved', { name: tab.name })}</p>
    <div class="actions">
      <button class="btn" type="button" onclick={onedit}>{t('dashboard.offer_edit')}</button>
      <button class="btn" type="button" onclick={onclose}>{t('dashboard.offer_close')}</button>
    </div>
  {:else}
    <p id="tab-rule-offer">{t('dashboard.offer_question', { name: tab.name })}</p>
    <div class="actions">
      {#each rulePresets(clientNetwork) as preset (preset)}
        <button class="btn" type="button" disabled={busy} onclick={() => choose(preset)}
          >{t(labels[preset], { network: clientNetwork })}</button
        >
      {/each}
      <button class="btn" type="button" onclick={onclose}>{t('dashboard.offer_no')}</button>
    </div>
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
</div>

<style>
  .offer {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-1) var(--space-2);
    margin: 0 0 var(--space-2);
    padding: var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
  .offer p {
    margin: 0;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }
  .error {
    flex-basis: 100%;
  }
</style>
