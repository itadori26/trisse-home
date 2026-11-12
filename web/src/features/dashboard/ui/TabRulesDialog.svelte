<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { MAX_TAB_RULES, parseNetworks, type Tab, type TabRule } from '../domain/dashboard';

  /** Props of the dialog that edits when a tab opens by default (#71). */
  interface Props {
    tab: Tab;
    /** Saves the rules; may throw to show an error. */
    onsubmit: (rules: TabRule[]) => Promise<void>;
    onclose: () => void;
  }
  let { tab, onsubmit, onclose }: Props = $props();

  /** Rule as edited: networks as free text. */
  interface Draft {
    days: number[];
    from: string;
    to: string;
    networks: string;
  }

  const weekdays = [1, 2, 3, 4, 5, 6, 0];
  // 8 November 2026 is a Sunday, so 8 + d is weekday d.
  const dayName = (d: number) =>
    new Date(2026, 10, 8 + d).toLocaleDateString(intlLocale(), { weekday: 'short' });

  // The dialog is recreated for every tab, so copying the rules once is intended.
  // svelte-ignore state_referenced_locally
  let drafts = $state<Draft[]>(
    tab.rules.map((r) => ({ days: [...r.days], from: r.from, to: r.to, networks: r.networks.join(', ') })),
  );
  let error = $state('');
  let busy = $state(false);

  function add() {
    drafts.push({ days: [], from: '', to: '', networks: '' });
  }

  function toggleDay(d: Draft, day: number, on: boolean) {
    d.days = on ? [...d.days, day] : d.days.filter((x) => x !== day);
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    busy = true;
    try {
      await onsubmit(
        drafts.map((d) => ({ days: d.days, from: d.from, to: d.to, networks: parseNetworks(d.networks) })),
      );
      onclose();
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={t('dashboard.rules_of', { name: tab.name })} {onclose}>
  <form onsubmit={submit}>
    <p class="muted small">{t('dashboard.rules_hint')}</p>
    {#each drafts as d, i (i)}
      <fieldset class="rule">
        <legend>{t('dashboard.rule_n', { n: i + 1 })}</legend>
        <div class="field">
          <span>{t('dashboard.rule_days')}</span>
          <div class="days">
            {#each weekdays as day (day)}
              <label
                ><input
                  type="checkbox"
                  checked={d.days.includes(day)}
                  onchange={(e) => toggleDay(d, day, e.currentTarget.checked)}
                />
                {dayName(day)}</label
              >
            {/each}
          </div>
        </div>
        <div class="field">
          <span>{t('dashboard.rule_time')}</span>
          <div class="times">
            <label
              >{t('dashboard.rule_from')} <input class="input time" type="time" bind:value={d.from} /></label
            >
            <label>{t('dashboard.rule_to')} <input class="input time" type="time" bind:value={d.to} /></label>
          </div>
        </div>
        <label class="field">
          <span>{t('dashboard.rule_networks')}</span>
          <input
            class="input"
            bind:value={d.networks}
            placeholder="192.168.178.0/24, 10.13.13.0/24"
            autocomplete="off"
            spellcheck="false"
          />
        </label>
        <button class="btn" type="button" onclick={() => drafts.splice(i, 1)}
          >{t('dashboard.remove_rule')}</button
        >
      </fieldset>
    {:else}
      <p class="muted">{t('dashboard.no_rules')}</p>
    {/each}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <p class="actions">
      <button class="btn" type="button" disabled={drafts.length >= MAX_TAB_RULES} onclick={add}
        >{t('dashboard.add_rule')}</button
      >
      <button class="btn btn-primary" type="submit" disabled={busy}>{t('dashboard.save')}</button>
    </p>
  </form>
</Modal>

<style>
  .rule {
    display: grid;
    gap: 0.5rem;
    margin: 0 0 0.75rem;
    justify-items: start;
  }
  .rule .field {
    width: 100%;
  }
  .days,
  .times {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem 0.75rem;
  }
  .time {
    width: auto;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    justify-content: flex-end;
  }
  .small {
    font-size: 0.85rem;
  }
</style>
