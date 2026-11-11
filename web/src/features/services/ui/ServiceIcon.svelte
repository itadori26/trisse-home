<script lang="ts">
  import { isImageIcon } from '../domain/service';

  /** Props of a service icon: an image URL, an emoji, or the first letter of the name. */
  interface Props {
    icon: string;
    name: string;
  }
  let { icon, name }: Props = $props();
</script>

{#if isImageIcon(icon)}
  <img src={icon} alt="" loading="lazy" referrerpolicy="no-referrer" />
{:else}
  <span class="glyph" aria-hidden="true">{icon || name.slice(0, 1).toUpperCase()}</span>
{/if}

<style>
  img,
  .glyph {
    width: clamp(24px, 1.8em, 48px);
    height: clamp(24px, 1.8em, 48px);
    flex: none;
  }
  img {
    object-fit: contain;
  }
  .glyph {
    display: grid;
    place-items: center;
    border-radius: 8px;
    background: var(--color-surface-2);
    font-weight: 600;
  }
</style>
