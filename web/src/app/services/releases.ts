import { ReleasesApi, ReleasesUseCases } from '../../features/releases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    releases: ReleasesUseCases;
  }
}

export const service = defineService('releases', (http) => new ReleasesUseCases(new ReleasesApi(http)));
