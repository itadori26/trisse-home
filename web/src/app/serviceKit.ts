/**
 * Building blocks of the service container: every feature registers its use
 * cases from its own src/app/services/<name>.ts (#193), so adding a feature
 * touches no shared file. A service file declares the field and how to create
 * it, side by side:
 *
 *     declare module '../serviceKit' {
 *       interface Services { media: MediaUseCases }
 *     }
 *     export const service = defineService('media', (http) => new MediaUseCases(new MediaApi(http)));
 */
import type { FetchHttpClient } from '../shared/api-client/http';

/** Use cases registered from src/app/services/*.ts; each file augments this interface. */
export interface Services {}

/** How one service is created. */
export interface ServiceFactory<K extends keyof Services = keyof Services> {
  key: K;
  create: (http: FetchHttpClient) => Services[K];
}

/** Type-checks that the created value matches the field the file declared. */
export function defineService<K extends keyof Services>(
  key: K,
  create: (http: FetchHttpClient) => Services[K],
): ServiceFactory<K> {
  return { key, create };
}
