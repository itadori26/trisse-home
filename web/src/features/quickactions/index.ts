/** Public API of the quick actions feature (#105, #67). */
export {
  type Prefixes,
  type QuickCommand,
  normaliseUrl,
  palettePrefixes,
  prefixesFromConfig,
  searchPrefixes,
} from './domain/command';
export {
  QuickActionError,
  QuickActionsUseCases,
  type QuickResult,
  type QuickTargets,
} from './usecase/quickActionsUseCases';
