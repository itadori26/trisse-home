<script lang="ts">
  import { type IconName, iconPaths } from './paths';

  /** Props of an inline SVG icon; decorative unless a label is given. */
  interface Props {
    name: IconName;
    size?: number;
    label?: string;
  }
  let { name, size = 24, label }: Props = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="1.75"
  stroke-linecap="round"
  stroke-linejoin="round"
  role={label ? 'img' : undefined}
  aria-label={label}
  aria-hidden={label ? undefined : 'true'}
>
  {#each iconPaths[name] as d (d)}<path {d} />{/each}
</svg>
