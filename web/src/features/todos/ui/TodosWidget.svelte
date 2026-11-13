<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import TodoEditDialog from './TodoEditDialog.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { describeReminder, describeRepeat, formatDue, parseTodosConfig, type Todo } from '../domain/todo';
  import type { TodosUseCases } from '../usecase/todosUseCases';

  /** Props of the to-do widget. */
  interface Props {
    todos: TodosUseCases;
    config: Record<string, unknown>;
  }
  let { todos, config }: Props = $props();

  const cfg = $derived(parseTodosConfig(config));
  let list = $state<Todo[]>([]);
  let text = $state('');
  let error = $state('');
  const hasDone = $derived(list.some((t) => t.done));
  let editing = $state<Todo | null>(null);
  const dayName = (d: number) =>
    new Date(2026, 10, 8 + d).toLocaleDateString(intlLocale(), { weekday: 'short' });

  async function run(action: () => Promise<unknown>) {
    error = '';
    try {
      await action();
      list = await todos.list(cfg.showDone);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    void cfg;
    run(async () => {});
  });

  function add(e: SubmitEvent) {
    e.preventDefault();
    if (text.trim() === '') return;
    run(() => todos.add(text)).then(() => (text = ''));
  }
</script>

<form onsubmit={add}>
  <input
    class="input"
    bind:value={text}
    placeholder={t('todos.add_a_to_do_2026_12')}
    aria-label={t('todos.new_to_do')}
  />
</form>
{#if error}<p class="error">{error}</p>{/if}
<ul class="list">
  {#each list as todo (todo.id)}
    <li class:done={todo.done}>
      <label>
        <input type="checkbox" checked={todo.done} onchange={() => run(() => todos.toggle(todo))} />
        <span class="title">{todo.title}</span>
      </label>
      {#if todo.repeat.kind}
        {@const text = describeRepeat(todo.repeat, dayName)}
        <span class="icon" title={text} aria-label={text} role="img">🔁</span>
      {/if}
      {#if todo.reminder.at}
        {@const text = describeReminder(todo.reminder)}
        <span class="icon" title={text} aria-label={text} role="img">🔔</span>
      {/if}
      {#if todo.due}<span class="due" class:overdue={todo.overdue}>{formatDue(todo.due, new Date())}</span
        >{/if}
      <button
        class="btn btn-icon del"
        type="button"
        aria-label={t('todos.edit', { title: todo.title })}
        onclick={() => (editing = todo)}>✎</button
      >
      <button
        class="btn btn-icon del"
        type="button"
        aria-label={t('todos.delete', { title: todo.title })}
        onclick={() => run(() => todos.delete(todo.id))}>✕</button
      >
    </li>
  {:else}
    <li class="muted">{t('todos.nothing_to_do')}</li>
  {/each}
</ul>
{#if editing}
  <TodoEditDialog
    todo={editing}
    {todos}
    onsaved={() => run(async () => {})}
    onclose={() => (editing = null)}
  />
{/if}
{#if hasDone}
  <button class="btn btn-icon muted" type="button" onclick={() => run(() => todos.clearCompleted())}
    >{t('todos.clear_completed')}</button
  >
{/if}

<style>
  form {
    margin-bottom: var(--space-2);
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 2px 0;
  }
  label {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex: 1;
    min-width: 0;
  }
  .title {
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .done .title {
    text-decoration: line-through;
    color: var(--color-text-muted);
  }
  .due {
    font-size: 0.8em;
    color: var(--color-text-muted);
    white-space: nowrap;
  }
  .icon {
    font-size: 0.75em;
  }
  .overdue {
    color: var(--color-danger);
  }
  .del {
    opacity: 0;
  }
  li:hover .del,
  .del:focus {
    opacity: 1;
  }
</style>
