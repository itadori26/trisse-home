import { registerFeatureMessages } from '../shared/i18n';

// Tells the i18n layer where the texts of the features live (#193). Imported
// by main.ts and the test setup before the first catalogue loads.
registerFeatureMessages({
  en: async () => (await import('./messages.en')).messages,
  de: async () => (await import('./messages.de')).messages,
});
