/**
 * Message keys contributed by features (#193). A feature declares its texts in
 * `features/<name>/i18n/en.ts` and `de.ts`, and `en.ts` adds its keys to this
 * interface, so `t('<key>')` stays type-checked without any shared file
 * listing the feature:
 *
 *     declare module '../../../shared/i18n/registry' {
 *       interface FeatureMessages extends Record<keyof typeof en, string> {}
 *     }
 *
 * src/app/featureMessages.ts collects the catalogues of all features.
 */
export interface FeatureMessages {}

/** One language of all features, keyed by message key. */
export type FeatureCatalogue = Record<string, string>;

/** Lazy loaders of the feature catalogues per language. */
export interface FeatureLoaders {
  en: () => Promise<FeatureCatalogue>;
  de: () => Promise<FeatureCatalogue>;
}
