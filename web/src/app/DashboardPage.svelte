<script lang="ts">
  import { DashboardView, type Widget, type WidgetConfig } from '../features/dashboard';
  import type { Container } from './container';
  import { definitionFor, widgetDefinitions } from './widgets';

  /** Props of the dashboard page. */
  interface Props {
    container: Container;
    editing: boolean;
    /** Widget kinds enabled in the backend. */
    widgetKinds: Set<string>;
  }
  let { container, editing, widgetKinds }: Props = $props();

  const types = $derived(widgetDefinitions.filter((d) => widgetKinds.has(d.type.kind)).map((d) => d.type));
</script>

<DashboardView
  uc={container.dashboard}
  {editing}
  widgetTypes={types}
  hasSettings={(kind) => !!definitionFor(kind)?.settings}
  chromeless={(kind) => !!definitionFor(kind)?.chromeless}
>
  {#snippet content(widget: Widget)}
    {@const def = definitionFor(widget.kind)}
    {#if def}
      {@const View = def.view.component}
      <View {...def.view.props(widget.config, container)} />
    {/if}
  {/snippet}
  {#snippet settings(widget: Widget, save: (config: WidgetConfig) => Promise<void>)}
    {@const def = definitionFor(widget.kind)}
    {#if def?.settings}
      {@const Settings = def.settings.component}
      <Settings {...def.settings.props(widget.config, container)} onsave={save} />
    {/if}
  {/snippet}
</DashboardView>
