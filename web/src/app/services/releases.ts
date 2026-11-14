import { ReleasesApi } from '../../features/releases/adapter/api/releasesApi';
import { ReleasesUseCases } from '../../features/releases/usecase/releasesUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    releases: ReleasesUseCases;
  }
}

// Direct imports keep the settings UI out of the initial bundle (#230).
export const service = defineService('releases', (http) => new ReleasesUseCases(new ReleasesApi(http)));
