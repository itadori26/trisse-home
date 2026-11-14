<script lang="ts">
  import { live } from './live.svelte';
</script>

<!-- The one polite live region of the app (#113); see announce(). -->
<div class="sr-only" aria-live="polite" aria-atomic="true">{live.message}</div>
