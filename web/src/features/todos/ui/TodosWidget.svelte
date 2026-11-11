<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { formatDue, parseTodosConfig, type Todo } from '../domain/todo';
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
    placeholder="Add a to-do… (@2026-12-24 sets a due date)"
    aria-label="New to-do"
  />
</form>
{#if error}<p class="error">{error}</p>{/if}
<ul class="list">
  {#each list as t (t.id)}
    <li class:done={t.done}>
      <label>
        <input type="checkbox" checked={t.done} onchange={() => run(() => todos.toggle(t))} />
        <span class="title">{t.title}</span>
      </label>
      {#if t.due}<span class="due" class:overdue={t.overdue}>{formatDue(t.due, new Date())}</span>{/if}
      <button
        class="btn btn-icon del"
        type="button"
        aria-label="Delete {t.title}"
        onclick={() => run(() => todos.delete(t.id))}>✕</button
      >
    </li>
  {:else}
    <li class="muted">Nothing to do 🎉</li>
  {/each}
</ul>
{#if hasDone}
  <button class="btn btn-icon muted" type="button" onclick={() => run(() => todos.clearCompleted())}
    >Clear completed</button
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
