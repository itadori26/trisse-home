/**
 * Composition root of the frontend: creates the HTTP client, the API adapters
 * and the use cases, and hands them to the UI. Nothing else instantiates
 * adapters, so features stay independent of each other.
 */
import { AuthApi, AuthUseCases } from '../features/auth';
import { FetchHttpClient } from '../shared/api-client/http';

/** All use cases available to the UI. */
export interface Container {
  http: FetchHttpClient;
  auth: AuthUseCases;
}

/** Builds the container. `onUnauthorized` fires when the session expired. */
export function createContainer(onUnauthorized: () => void): Container {
  const http = new FetchHttpClient(onUnauthorized);
  return {
    http,
    auth: new AuthUseCases(new AuthApi(http), http),
  };
}
