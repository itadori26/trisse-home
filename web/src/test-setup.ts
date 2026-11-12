import { loadLanguage } from './shared/i18n';

// The UI catalogues are loaded lazily (#116); load the active one (English
// in Node) before any test calls t(), as App.svelte does before rendering.
await loadLanguage();
