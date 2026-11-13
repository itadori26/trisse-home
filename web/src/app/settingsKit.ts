/**
 * Building blocks of the settings dialog: every feature adds its section from
 * its own src/app/settings/<name>.ts (#193), so the dialog needs no edit.
 */
import type { Component } from 'svelte';
import type { MessageKey } from '../shared/i18n';
import type { Container } from './container';

/** What the dialog knows besides the container. */
export interface SettingsContext {
  /** Enabled backend modules. */
  modules: ReadonlySet<string>;
}

/** One section of the settings dialog, with its component loaded on demand. */
export interface SettingsSection {
  id: string;
  label: MessageKey;
  /** Shown only while this backend module is enabled. */
  module?: string;
  load: () => Promise<Component<Record<string, unknown>>>;
  props: (container: Container, context: SettingsContext) => Record<string, unknown>;
}

/**
 * Type-checks the props against the component and erases the type for the
 * registry (the same approach as lazy() in widgetKit.ts).
 */
export function defineSettingsSection<P extends object>(def: {
  id: string;
  label: MessageKey;
  module?: string;
  load: () => Promise<Component<P>>;
  props: (container: Container, context: SettingsContext) => P;
}): SettingsSection {
  return def as unknown as SettingsSection;
}
