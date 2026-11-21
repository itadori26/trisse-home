<script lang="ts">
  import { intlLocale, type MessageKey, t, textsIn } from '../shared/i18n';
  import { isTyping } from '../shared/shortcuts/registry';
  import { errorMessage } from '../shared/api-client/http';
  import { tick } from 'svelte';
  import { areaOf, groupByArea, moveFocus, type SettingsArea } from './settingsNav';
  import { registeredSections } from './settingsSections';
  import { type CatalogSection, OWN_SECTIONS, visibleSections } from './settingsCatalog';
  import { type AreaStatus, loadAreaStatuses, type StatusSources } from './settingsStatus';
  import { advancedProblem } from './settingsAdvancedHint';
  import { attention } from './attention.svelte';
  import { loadRecent, rememberChange } from './settingsRecent';
  import type { Appearance } from '../features/appearance';
  import Modal from '../shared/ui/Modal.svelte';
  import SettingsHome from './SettingsHome.svelte';
  import SettingsSearch from './SettingsSearch.svelte';
  import { buildIndex, type SettingsHit } from './settingsSearch';
  import { highlightField } from './settingsHighlight';
  import SettingsSectionView from './SettingsSectionView.svelte';
  import type { Container } from './container';

  /** Props of the global settings dialog. */
  interface Props {
    container: Container;
    modules: Set<string>;
    /** Who owns each module's data (ADR 0018); members see personal ones. */
    dataKinds?: Record<string, string>;
    appearance: Appearance;
    onappearance: (a: Appearance) => void;
    /** Restarts the guided tour; omitted when the onboarding module is disabled. */
    onrestarttour?: () => void;
    /** Section to open first, e.g. "appearance" from the launcher (#261). */
    section?: string;
    onclose: () => void;
  }
  let {
    container,
    modules,
    dataKinds = {},
    appearance,
    onappearance,
    onrestarttour,
    section,
    onclose,
  }: Props = $props();

  type Section = CatalogSection;
  /** Display name of a section: a plugin's own label or the translated key. */
  const nameOf = (s: Section) => s.title?.() ?? t(s.label);
  const sections: Section[] = [...OWN_SECTIONS, ...registeredSections];

  /**
   * Advanced settings (#602) are listed while the account switched them on.
   * A link straight to an advanced section (docs, search command) shows
   * them for this visit without changing the stored choice.
   */
  let visitAdvanced = $state(sections.some((s) => s.id === section && s.advanced));
  const advanced = $derived(!!appearance.showAdvanced || visitAdvanced);
  const viewer = $derived({
    modules,
    dataKinds,
    profile: appearance.profile,
    localMode: container.localMode,
    guest: container.auth.guest,
    member: container.auth.member,
    admin: container.auth.admin,
    owner: container.auth.owner,
  });
  const visible = $derived(visibleSections(sections, { ...viewer, advanced }));
  /** The switch is offered only to accounts that have something advanced to show and may save it. */
  const canSwitchAdvanced = $derived(
    !container.auth.guest &&
      !container.auth.kiosk &&
      visibleSections(sections, { ...viewer, advanced: true }).some(
        (s) => s.advanced || (s.advancedFields?.length ?? 0) > 0,
      ),
  );
  /**
   * A hidden advanced section whose connection fails (#622): the start page
   * points at it so an admin does not miss the problem behind the switch.
   */
  const advancedAttention = $derived(
    canSwitchAdvanced
      ? advancedProblem({
          items: attention.items,
          sections: visibleSections(sections, { ...viewer, advanced: true }),
          showAdvanced: !!appearance.showAdvanced,
          admin: container.auth.admin,
        })
      : undefined,
  );
  /** Opens the section for this visit; the stored switch stays off. */
  function openAdvancedAttention(id: string) {
    visitAdvanced = true;
    void go({ kind: 'section', id });
  }
  let advancedError = $state('');
  let savingAdvanced = $state(false);
  async function setAdvanced(on: boolean) {
    savingAdvanced = true;
    advancedError = '';
    try {
      onappearance(await container.appearance.save({ ...appearance, showAdvanced: on }));
      visitAdvanced = false;
    } catch (err) {
      advancedError = errorMessage(err);
    } finally {
      savingAdvanced = false;
    }
  }
  const areas = $derived(groupByArea(visible, nameOf, intlLocale()));

  /**
   * Where the dialog is (#600): the start page, one area's list of sections,
   * or a section. A deep link (?settings=<id>, search commands) opens its
   * section directly; an unknown or hidden one falls back to the start page.
   */
  type View = { kind: 'home' } | { kind: 'area'; area: SettingsArea } | { kind: 'section'; id: string };
  // svelte-ignore state_referenced_locally
  let view = $state<View>(section ? { kind: 'section', id: section } : { kind: 'home' });
  const current = $derived.by((): View => {
    if (view.kind === 'section') {
      const id = view.id;
      return visible.some((s) => s.id === id) ? view : { kind: 'home' };
    }
    if (view.kind === 'area') {
      const area = view.area;
      return areas.some((a) => a.id === area) ? view : { kind: 'home' };
    }
    return view;
  });
  const currentSection = $derived(
    current.kind === 'section' ? visible.find((s) => s.id === current.id) : undefined,
  );
  const currentArea = $derived.by(() => {
    if (current.kind === 'area') return areas.find((a) => a.id === current.area);
    if (currentSection) return areas.find((a) => a.id === areaOf(currentSection));
    return undefined;
  });
  /** An area with one section opens that section; a list of one adds a click. */
  const single = (a: { sections: readonly Section[] }) => a.sections.length === 1;
  /**
   * The navigation entry that stands for the current view: a section in the
   * open area's sub-list, else its area, else the start page. It is the one
   * entry in the tab order (roving tabindex), so Tab reaches the list once.
   */
  const navActive = $derived(
    current.kind === 'section' && currentArea && !single(currentArea)
      ? `section:${current.id}`
      : currentArea
        ? `area:${currentArea.id}`
        : 'home',
  );

  let nav = $state<HTMLElement>();
  let content = $state<HTMLElement>();
  let heading = $state<HTMLElement>();

  /** Shows a view and moves the focus to its heading for screen readers. */
  async function go(next: View) {
    view = next;
    if (content) content.scrollTop = 0;
    await tick();
    heading?.focus();
  }

  function openArea(id: SettingsArea) {
    const a = areas.find((x) => x.id === id);
    if (!a) return;
    const only = a.sections[0];
    void go(single(a) && only ? { kind: 'section', id: only.id } : { kind: 'area', area: id });
  }
  const openSection = (id: string) => void go({ kind: 'section', id });

  /** One level up: section → its area (unless it is the area's only one) → start page. */
  function up() {
    if (current.kind === 'section' && currentArea && !single(currentArea)) {
      void go({ kind: 'area', area: currentArea.id });
    } else {
      void go({ kind: 'home' });
    }
  }

  /** Arrow keys, Home and End move between the entries; Enter selects. */
  function navKey(e: KeyboardEvent) {
    const buttons = [...(nav?.querySelectorAll<HTMLButtonElement>('button[data-nav]') ?? [])];
    const next = moveFocus(e.key, buttons.indexOf(e.target as HTMLButtonElement), buttons.length);
    if (next === null) return;
    e.preventDefault();
    buttons[next]?.focus();
  }

  /** Sections of the connections area that are not an integration of their own. */
  const NOT_INTEGRATIONS = new Set(['tokens', 'attention', 'connections']);
  // Statuses of the area tiles: each source only for a viewer who may read it.
  let statuses = $state<Partial<Record<SettingsArea, AreaStatus>>>({});
  $effect(() => {
    const shown = (id: string) => visible.some((s) => s.id === id);
    const auth = container.auth;
    const sources: StatusSources = {
      twoFactor:
        shown('account') && !auth.kiosk && !container.localMode ? () => auth.twoFactorStatus() : undefined,
      channels: shown('notifications') ? () => container.notifications.channels() : undefined,
      backups: shown('backups') ? () => container.backup.status() : undefined,
      version: shown('about') ? () => container.system.version() : undefined,
      connections: areas
        .find((a) => a.id === 'connections')
        ?.sections.filter((s) => !NOT_INTEGRATIONS.has(s.id)).length,
    };
    const ids = areas.map((a) => a.id);
    let stale = false;
    loadAreaStatuses(ids, sources).then((s) => {
      if (!stale) statuses = s;
    });
    return () => (stale = true);
  });

  // "Recently changed" (#600): a change or a submitted form inside a section
  // counts, so the list needs no hook in every feature's settings form.
  let recentIds = $state(loadRecent());
  const recent = $derived(
    recentIds.flatMap((id) => {
      const s = visible.find((x) => x.id === id);
      const area = s && areas.find((a) => a.id === areaOf(s));
      return s && area ? [{ id, name: nameOf(s), area: t(area.label) }] : [];
    }),
  );
  function noteChange() {
    if (current.kind === 'section') recentIds = rememberChange(current.id);
  }

  // The settings search (#601) finds areas, sections and fields in English
  // and German whatever the UI language; the catalogues load on first use.
  let otherTexts = $state<((key: MessageKey) => string)[]>([]);
  async function loadSearchTexts() {
    try {
      otherTexts = await Promise.all([textsIn('en'), textsIn('de')]);
    } catch {
      // Offline: the search still works in the UI language.
    }
  }
  const index = $derived(
    buildIndex(areas, {
      ui: t,
      others: otherTexts,
      sectionName: nameOf,
      // A plugin's section brings its own labels through keywords.
      sectionLabel: (s) => (s.title ? undefined : s.label),
      advanced,
    }),
  );
  let search = $state<{ focus: () => void }>();

  async function pick(hit: SettingsHit) {
    if (hit.kind === 'area') return openArea(hit.area);
    await go({ kind: 'section', id: hit.section });
    if (hit.field && content) await highlightField(content, hit.field);
  }

  /** "/" focuses the search from anywhere in the dialog, as on the dashboard. */
  async function slash(e: KeyboardEvent) {
    if (e.key !== '/' || e.ctrlKey || e.metaKey || e.altKey || isTyping(e.target)) return;
    e.preventDefault();
    if (current.kind !== 'home') {
      view = { kind: 'home' };
      await tick();
    }
    search?.focus();
  }

  const title = $derived(
    currentSection ? nameOf(currentSection) : currentArea ? t(currentArea.label) : t('settings.overview'),
  );
</script>

<Modal title={t('app.settings')} size="wide" {onclose}>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="layout" onkeydown={slash}>
    <nav class="areas" aria-label={t('settings.areas')} bind:this={nav}>
      <button
        type="button"
        data-nav="home"
        class:current={navActive === 'home'}
        aria-current={navActive === 'home' ? 'page' : undefined}
        tabindex={navActive === 'home' ? 0 : -1}
        onkeydown={navKey}
        onclick={() => go({ kind: 'home' })}>{t('settings.overview')}</button
      >
      {#each areas as a (a.id)}
        {@const here = currentArea?.id === a.id}
        <button
          type="button"
          data-nav={a.id}
          class:current={navActive === `area:${a.id}`}
          class:open={here}
          aria-current={navActive === `area:${a.id}` ? 'page' : undefined}
          tabindex={navActive === `area:${a.id}` ? 0 : -1}
          onkeydown={navKey}
          onclick={() => openArea(a.id)}>{t(a.label)}</button
        >
        {#if here && !single(a)}
          {#each a.sections as s (s.id)}
            <button
              type="button"
              class="sub"
              data-nav={s.id}
              data-section={s.id}
              class:current={navActive === `section:${s.id}`}
              aria-current={navActive === `section:${s.id}` ? 'page' : undefined}
              tabindex={navActive === `section:${s.id}` ? 0 : -1}
              onkeydown={navKey}
              onclick={() => openSection(s.id)}>{nameOf(s)}</button
            >
          {/each}
        {/if}
      {/each}
    </nav>
    <section
      class="content"
      bind:this={content}
      aria-label={title}
      onchange={noteChange}
      onsubmit={noteChange}
    >
      {#if current.kind !== 'home'}
        <button class="btn back" type="button" onclick={up}>
          ← {current.kind === 'section' && currentArea && !single(currentArea)
            ? t(currentArea.label)
            : t('settings.overview')}
        </button>
      {/if}
      <h3 class="title" tabindex="-1" bind:this={heading}>{title}</h3>
      {#if current.kind === 'home'}
        <SettingsSearch bind:this={search} {index} onfirstuse={loadSearchTexts} onpick={pick} />
        <SettingsHome {areas} {statuses} {recent} onarea={openArea} onsection={openSection} />
        {#if canSwitchAdvanced}
          <div class="advanced">
            <label class="switch">
              <input
                type="checkbox"
                checked={!!appearance.showAdvanced}
                disabled={savingAdvanced}
                aria-describedby="settings-advanced-hint"
                onchange={(e) => setAdvanced(e.currentTarget.checked)}
              />
              <span>{t('settings.advanced.label')}</span>
            </label>
            <p class="muted" id="settings-advanced-hint">{t('settings.advanced.hint')}</p>
            {#if advancedAttention}
              {@const target = advancedAttention}
              <p class="advanced-attention">
                <span class="dot" aria-hidden="true"></span>
                {t('settings.advanced.attention')}
                <button type="button" class="link" onclick={() => openAdvancedAttention(target.id)}>
                  {t('settings.advanced.attention_open', { name: nameOf(target) })}
                </button>
              </p>
            {/if}
            {#if advancedError}<p class="error" role="alert">{advancedError}</p>{/if}
          </div>
        {/if}
      {:else if current.kind === 'area' && currentArea}
        <p class="muted">{t(currentArea.hint)}</p>
        <ul class="list">
          {#each currentArea.sections as s (s.id)}
            <li>
              <button type="button" data-section={s.id} onclick={() => openSection(s.id)}>
                <span>{nameOf(s)}</span><span aria-hidden="true">›</span>
              </button>
            </li>
          {/each}
        </ul>
      {:else if current.kind === 'section'}
        <SettingsSectionView
          id={current.id}
          {container}
          {modules}
          {appearance}
          {onappearance}
          {onrestarttour}
          {advanced}
        />
      {/if}
    </section>
  </div>
</Modal>

<style>
  /* Two columns: the areas stay, the content scrolls on its own (#261, #600). */
  .layout {
    display: grid;
    grid-template-columns: minmax(180px, 230px) 1fr;
    gap: var(--space-4);
    height: min(72vh, 760px);
  }
  .areas,
  .content {
    overflow-y: auto;
    min-height: 0;
    overscroll-behavior: contain;
  }
  .areas {
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--color-border);
    padding-right: var(--space-2);
  }
  .content {
    min-width: 0;
    padding-right: var(--space-1);
  }
  .areas button,
  .list button {
    text-align: left;
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    padding: 6px var(--space-2);
    border-radius: var(--radius-small);
    cursor: pointer;
  }
  .areas button.open {
    font-weight: 600;
  }
  .areas button.sub {
    padding-left: var(--space-6);
    font-size: 0.92rem;
  }
  .areas button:hover,
  .list button:hover {
    background: var(--color-surface-2);
  }
  .areas button.current {
    background: color-mix(in srgb, var(--color-accent) 18%, transparent);
    font-weight: 600;
  }
  .areas button:focus-visible,
  .list button:focus-visible {
    outline: 2px solid var(--color-accent-text);
    outline-offset: -2px;
  }
  .title {
    margin: 0 0 var(--space-3);
  }
  .title:focus {
    outline: none;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
  .list li + li {
    border-top: 1px solid var(--color-border);
  }
  .list button {
    display: flex;
    justify-content: space-between;
    width: 100%;
    padding: 10px var(--space-3);
  }
  .back {
    display: none;
  }
  .advanced {
    margin-top: var(--space-6);
    padding-top: var(--space-3);
    border-top: 1px solid var(--color-border);
  }
  .advanced p {
    margin: var(--space-1) 0 0;
    font-size: 0.85rem;
  }
  .advanced-attention {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
  }
  .advanced-attention .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-warning);
  }
  .advanced-attention .link {
    min-height: 44px;
    padding: 0;
    border: 0;
    background: none;
    color: var(--color-accent-text);
    font: inherit;
    text-decoration: underline;
    cursor: pointer;
  }
  .switch {
    display: inline-flex;
    gap: var(--space-2);
    align-items: center;
    cursor: pointer;
  }
  /* A field found by the search (#601), marked for a moment. */
  .content :global(.settings-found) {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 4px;
    border-radius: var(--radius-small);
    transition: outline-color 0.3s;
  }
  /* Phones: one page at a time with a back link instead of the area column. */
  @media (max-width: 640px) {
    .layout {
      display: block;
      height: auto;
    }
    .areas {
      display: none;
    }
    .back {
      display: inline-flex;
      margin-bottom: var(--space-3);
    }
  }
</style>
