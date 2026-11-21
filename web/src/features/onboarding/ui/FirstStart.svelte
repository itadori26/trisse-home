<script lang="ts">
  import type { Snippet } from 'svelte';
  import { t, type MessageKey } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { browserLocation } from '../../../shared/places/places';
  import type { ThemeMode } from '../../../shared/theme/theme';
  import { firstStartSteps, type FirstStartStep, type Personal, type Place } from '../domain/firstStart';

  /**
   * Props of the first-start flow (#671): one question per screen, every one
   * skippable, before the template choice and the tour. It writes the same
   * settings as the settings pages, through the ports it is given.
   */
  interface Props {
    /** Invited members get the personal steps only (name, language, look). */
    owner: boolean;
    /** The personal settings as they are now; undefined hides those steps. */
    personal?: Personal;
    /** Saves changed personal settings. */
    savePersonal: (patch: Partial<Personal>) => Promise<void>;
    /** Shows a change right away (language, look); null undoes the preview. */
    preview: (patch: Partial<Personal> | null) => void;
    /** Saved places (#591); undefined hides the place step. */
    places?: {
      /** Place search (the weather geocoder); omitted when it is off. */
      search?: (query: string) => Promise<Omit<Place, 'id'>[]>;
      save: (place: Omit<Place, 'id'>) => Promise<Place>;
    };
    /** Further steps from plugins, e.g. the instance name (owner only). */
    extraSteps?: { id: string; title: string }[];
    /** Draws a plugin step; it registers what Next saves. */
    extra?: Snippet<[string, (save: () => Promise<void>) => void]>;
    /** The flow is over; place is the one saved on the way, if any. */
    ondone: (result: { place?: Place }) => void;
  }
  let { owner, personal, savePersonal, preview, places, extraSteps = [], extra, ondone }: Props = $props();

  // The steps are fixed when the flow opens, so the progress does not jump.
  // svelte-ignore state_referenced_locally
  const steps: FirstStartStep[] = firstStartSteps({
    owner,
    personal: personal !== undefined,
    look: personal !== undefined && !personal.themeLocked,
    places: places !== undefined,
    extra: extraSteps.map((s) => s.id),
  });
  let index = $state(0);
  const step = $derived(steps[index]);
  let busy = $state(false);
  let error = $state('');

  // Drafts of the answers, prefilled with what is stored.
  // svelte-ignore state_referenced_locally
  let name = $state(personal?.greetingName ?? '');
  // svelte-ignore state_referenced_locally
  let language = $state(personal?.language ?? '');
  // svelte-ignore state_referenced_locally
  let theme = $state<ThemeMode>(personal?.theme ?? 'system');
  // svelte-ignore state_referenced_locally
  let accent = $state(personal?.accentColor ?? '');
  let query = $state('');
  let results = $state<Omit<Place, 'id'>[]>([]);
  let chosen = $state<Omit<Place, 'id'> | undefined>();
  let saved = $state<Place | undefined>();
  let submit: (() => Promise<void>) | undefined;

  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  function titleOf(s: FirstStartStep | undefined): string {
    if (!s) return '';
    if (s.startsWith('extra:')) return extraSteps.find((x) => `extra:${x.id}` === s)?.title ?? '';
    return t(`first_start.${s}.title` as MessageKey);
  }

  function go(next: number) {
    error = '';
    submit = undefined;
    if (next >= steps.length) {
      preview(null);
      ondone({ ...(saved ? { place: saved } : {}) });
    } else index = next;
  }

  /** Saves the answer of the current step, then moves on. */
  async function save() {
    busy = true;
    error = '';
    try {
      if (step === 'name' && name.trim() !== (personal?.greetingName ?? '')) {
        await savePersonal({ greetingName: name.trim() });
      } else if (step === 'language' && language !== personal?.language) {
        await savePersonal({ language });
      } else if (step === 'look' && (theme !== personal?.theme || accent !== personal?.accentColor)) {
        await savePersonal({ theme, accentColor: accent });
      } else if (step === 'place' && chosen && places) {
        saved = await places.save({ ...chosen, name: t('first_start.place.name') });
      } else if (step?.startsWith('extra:')) {
        await submit?.();
      }
      go(index + 1);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  /** Leaves the current step unanswered; a preview of it is undone. */
  function skip() {
    if (step === 'language') language = personal?.language ?? '';
    if (step === 'look') {
      theme = personal?.theme ?? 'system';
      accent = personal?.accentColor ?? '';
    }
    preview(previewPatch());
    go(index + 1);
  }

  // Language and look show at once; the stored values come back on skip.
  function previewPatch(): Partial<Personal> {
    return { language, theme, accentColor: accent };
  }
  $effect(() => preview(previewPatch()));

  async function find(e: SubmitEvent) {
    e.preventDefault();
    if (!places?.search) return;
    error = '';
    try {
      results = await places.search(query);
      if (results.length === 0) error = t('templates.no_results');
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function locate() {
    error = '';
    try {
      const at = await browserLocation();
      chosen = { name: '', label: t('first_start.place.here'), ...at };
    } catch {
      error = t('templates.location_denied');
    }
  }
</script>

<Modal title={titleOf(step)} onclose={() => go(steps.length)}>
  <p class="muted small progress">{t('first_start.progress', { n: index + 1, total: steps.length })}</p>
  {#if step === 'name'}
    <label class="field">
      <span>{t('first_start.name.label')}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={name} maxlength="40" autocomplete="given-name" autofocus />
    </label>
    <p class="muted small">{t('first_start.name.hint')}</p>
  {:else if step === 'place'}
    <p class="muted">{t('first_start.place.intro')}</p>
    {#if places?.search}
      <form class="row" onsubmit={find}>
        <input class="input grow" bind:value={query} aria-label={t('templates.location_search')} />
        <button class="btn" type="submit">{t('templates.search')}</button>
      </form>
      {#if results.length}
        <ul class="results" aria-label={t('first_start.place.results')}>
          {#each results as r (`${r.lat},${r.lon}`)}
            <li>
              <button
                class="btn"
                type="button"
                aria-pressed={chosen?.lat === r.lat && chosen?.lon === r.lon}
                onclick={() => (chosen = r)}>{r.label}</button
              >
            </li>
          {/each}
        </ul>
      {/if}
    {/if}
    <button class="btn" type="button" onclick={locate}>{t('templates.my_location')}</button>
    {#if chosen}<p role="status">{t('first_start.place.chosen', { place: chosen.label })}</p>{/if}
  {:else if step === 'language'}
    <label class="field">
      <span>{t('appearance.language')}</span>
      <select class="input" bind:value={language}>
        <option value="">{t('appearance.language_browser')}</option>
        <option value="en">English</option>
        <option value="de">Deutsch</option>
      </select>
    </label>
    <p class="muted small">{t('first_start.language.time_zone', { zone: timeZone })}</p>
  {:else if step === 'look'}
    <fieldset class="field">
      <legend>{t('appearance.theme')}</legend>
      {#each ['system', 'light', 'dark'] as const as mode (mode)}
        <label class="choice">
          <input type="radio" name="theme" value={mode} bind:group={theme} />
          {t(`appearance.${mode}`)}
        </label>
      {/each}
    </fieldset>
    <label class="field">
      <span>{t('appearance.accent_color')}</span>
      <input type="color" value={accent || '#3365e8'} oninput={(e) => (accent = e.currentTarget.value)} />
    </label>
    <p class="muted small">{t('first_start.look.hint')}</p>
  {:else if step?.startsWith('extra:') && extra}
    {@render extra(step.slice('extra:'.length), (fn) => (submit = fn))}
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <div class="actions">
    {#if index > 0}
      <button class="btn" type="button" disabled={busy} onclick={() => go(index - 1)}
        >{t('onboarding.back')}</button
      >
    {/if}
    <span class="grow"></span>
    <button class="btn" type="button" disabled={busy} onclick={skip}>{t('first_start.skip')}</button>
    <button class="btn btn-primary" type="button" disabled={busy} onclick={save}
      >{t('first_start.next')}</button
    >
  </div>
</Modal>

<style>
  .progress {
    margin-top: 0;
  }
  .small {
    font-size: 0.85rem;
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    margin: 0 0 var(--space-3);
    border: 0;
    padding: 0;
  }
  .choice {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .row,
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
  }
  .row {
    margin-bottom: var(--space-2);
  }
  .actions {
    margin-top: var(--space-3);
  }
  .grow {
    flex: 1;
    min-width: 8rem;
  }
  .results {
    list-style: none;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    margin: 0 0 var(--space-2);
  }
</style>
