<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { type GiftIdea, groupByPerson, type NextBirthday } from '../domain/gifts';
  import type { BirthdaySource, GiftsUseCases } from '../usecase/giftsUseCases';
  import GiftDialog from './GiftDialog.svelte';

  /** Props of the gift ideas widget (#684). */
  interface Props {
    gifts: GiftsUseCases;
    birthdays: BirthdaySource;
  }
  let { gifts, birthdays }: Props = $props();
  const poll = usePolling();

  let ideas = $state<GiftIdea[]>([]);
  let upcoming = $state<NextBirthday[]>([]);
  let error = $state('');
  let editing = $state<{ idea: GiftIdea | null; person: string } | null>(null);

  const groups = $derived(groupByPerson(ideas, upcoming));
  const people = $derived([...new Set([...groups.map((g) => g.person), ...upcoming.map((b) => b.name)])]);

  async function load() {
    try {
      ideas = await gifts.list();
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
    upcoming = await birthdays.upcoming().catch(() => []);
  }

  $effect(() => poll(load, 30 * 60_000));

  function birthdayLabel(b: NextBirthday): string {
    const date = new Date(`${b.next}T12:00:00`).toLocaleDateString(intlLocale(), {
      day: 'numeric',
      month: 'short',
    });
    if (b.inDays === 0) return t('gifts.birthday_today');
    return t('gifts.birthday_in', { date, n: b.inDays });
  }

  async function toggle(idea: GiftIdea) {
    try {
      await gifts.setBought(idea, !idea.bought);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove(idea: GiftIdea) {
    if (!confirm(t('common.confirm_delete', { name: idea.title }))) return;
    try {
      await gifts.remove(idea.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if groups.length === 0}
  <p class="muted">{t('gifts.empty')}</p>
{/if}
{#each groups as g (g.person)}
  <section>
    <h3>
      <span class="person">{g.person}</span>
      {#if g.birthday}<span class="birthday muted">{birthdayLabel(g.birthday)}</span>{/if}
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('gifts.add_for', { person: g.person })}
        onclick={() => (editing = { idea: null, person: g.person })}>+</button
      >
    </h3>
    <ul>
      {#each g.ideas as idea (idea.id)}
        <li class:bought={idea.bought}>
          <input
            type="checkbox"
            checked={idea.bought}
            aria-label={t('gifts.bought_named', { title: idea.title })}
            onchange={() => toggle(idea)}
          />
          <span class="title">
            {#if idea.link}
              <a href={idea.link} target="_blank" rel="noopener noreferrer">{idea.title}</a>
            {:else}
              {idea.title}
            {/if}
            {#if idea.note}<span class="note muted">{idea.note}</span>{/if}
          </span>
          <button
            class="btn btn-icon"
            type="button"
            aria-label={t('gifts.edit_named', { title: idea.title })}
            onclick={() => (editing = { idea, person: idea.person })}>✎</button
          >
          <button
            class="btn btn-icon"
            type="button"
            aria-label={t('gifts.remove', { title: idea.title })}
            onclick={() => remove(idea)}>✕</button
          >
        </li>
      {/each}
    </ul>
  </section>
{/each}
<button class="btn add" type="button" onclick={() => (editing = { idea: null, person: '' })}
  >+ {t('gifts.new')}</button
>

{#if editing}
  <GiftDialog
    {gifts}
    idea={editing.idea}
    person={editing.person}
    {people}
    onsaved={() => load().catch(() => {})}
    onclose={() => (editing = null)}
  />
{/if}

<style>
  section {
    margin-bottom: var(--space-2);
  }
  h3 {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: 0.9rem;
    margin: 0;
  }
  .person {
    font-weight: 600;
  }
  .birthday {
    flex: 1;
    font-size: 0.8rem;
    font-weight: normal;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-width: 0;
  }
  li input[type='checkbox'] {
    width: 1.25rem;
    height: 1.25rem;
    margin: 0.75rem;
    flex-shrink: 0;
  }
  .title {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }
  .note {
    font-size: 0.8rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  li.bought .title {
    text-decoration: line-through;
    opacity: 0.6;
  }
</style>
