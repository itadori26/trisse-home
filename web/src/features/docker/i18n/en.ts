/** English UI texts added to the docker feature (#119, #193). */
export const en = {
  'docker.only_stopped': 'Only show stopped containers',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
