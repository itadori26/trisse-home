<script lang="ts">
  import type { Snippet } from 'svelte';

  /** Props of the centered card used by the setup and login screens. */
  interface Props {
    title: string;
    subtitle?: string;
    children: Snippet;
  }
  let { title, subtitle, children }: Props = $props();
</script>

<main>
  <section class="card">
    <h1>{title}</h1>
    {#if subtitle}<p class="muted">{subtitle}</p>{/if}
    {@render children()}
  </section>
</main>

<style>
  main {
    min-height: 100vh;
    display: grid;
    place-items: center;
    padding: var(--space-4);
  }
  section {
    width: min(420px, 100%);
    padding: var(--space-6);
  }
  h1 {
    margin: 0 0 var(--space-2);
    font-size: 1.4rem;
  }
</style>
