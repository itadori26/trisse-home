/**
 * Building blocks of the widget registry: the definition type and lazy().
 * Every widget is defined in its own src/app/widgets/<kind>.ts (#193).
 */
import type { Component } from 'svelte';
import type { WidgetType } from '../features/dashboard';
import type { IconName } from '../shared/ui/icons/paths';
import type { Container } from './container';

/** Configuration object stored per widget. */
export type Config = Record<string, unknown>;

/** A lazily loaded component together with a function that builds its props (#116). */
export interface Renderer {
  /** Loads the component's chunk; the promise is cached per kind. */
  load: () => Promise<Component<Record<string, unknown>>>;
  props: (config: Config, c: Container) => Record<string, unknown>;
}

/** Everything the dashboard needs to know about one widget kind. */
export interface WidgetDefinition {
  type: WidgetType;
  /** Tile icon of the widget picker (#104); required for every kind. */
  icon: IconName;
  /** Rendered without card chrome outside edit mode. */
  chromeless?: boolean;
  view: Renderer;
  /** Settings form; it additionally receives an `onsave(config)` prop. */
  settings?: Renderer;
}

/**
 * Type-checks props against the lazily imported component and erases the
 * type for the registry. Only this file knows how widget code is loaded.
 */
export function lazy<P extends object>(
  load: () => Promise<{ default: Component<P> }>,
  props: (config: Config, c: Container) => Omit<P, 'onsave'>,
): Renderer {
  let loaded: Promise<Component<Record<string, unknown>>> | undefined;
  return {
    load: () => {
      loaded ??= load().then((m) => m.default as unknown as Component<Record<string, unknown>>);
      // A failed chunk (e.g. offline after an update) can be retried.
      loaded.catch(() => (loaded = undefined));
      return loaded;
    },
    props,
  };
}
